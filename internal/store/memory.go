package store

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ryoshimaru/MindMap/internal/domain"
)

var (
	ErrConflict     = errors.New("conflict")
	ErrInvalidAuth  = errors.New("invalid auth")
	ErrInvalidAIKey = errors.New("invalid AI key")
	ErrInvalidState = errors.New("invalid state")
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrValidation   = errors.New("validation")
)

type MemoryStore struct {
	mu sync.RWMutex

	users        map[string]*domain.UserRecord
	usersByEmail map[string]string
	sessions     map[string]string
	aiSettings   map[string]*domain.AISettingsRequest
	profiles     map[string]*domain.UserProfile

	goals       map[string]*domain.GoalRecord
	goalOrder   []string
	goalContext map[string]*domain.GoalContext

	tasks      map[string]*domain.Task
	taskOrder  []string
	deps       map[string]*domain.TaskDependency
	depsByTask map[string][]string

	versions       map[string]*domain.PlanVersion
	versionsByGoal map[string][]string

	generations       map[string]*domain.GenerationDetails
	generationsByGoal map[string][]string
	feedback          map[string]*domain.Feedback
	feedbackByGen     map[string][]string

	planningRequests map[string]*domain.PlanningRequest

	deletedTasksByGoal map[string]int
}

func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		users:              make(map[string]*domain.UserRecord),
		usersByEmail:       make(map[string]string),
		sessions:           make(map[string]string),
		aiSettings:         make(map[string]*domain.AISettingsRequest),
		profiles:           make(map[string]*domain.UserProfile),
		goals:              make(map[string]*domain.GoalRecord),
		goalContext:        make(map[string]*domain.GoalContext),
		tasks:              make(map[string]*domain.Task),
		deps:               make(map[string]*domain.TaskDependency),
		depsByTask:         make(map[string][]string),
		versions:           make(map[string]*domain.PlanVersion),
		versionsByGoal:     make(map[string][]string),
		generations:        make(map[string]*domain.GenerationDetails),
		generationsByGoal:  make(map[string][]string),
		feedback:           make(map[string]*domain.Feedback),
		feedbackByGen:      make(map[string][]string),
		planningRequests:   make(map[string]*domain.PlanningRequest),
		deletedTasksByGoal: make(map[string]int),
	}

	store.seed()
	return store
}

func (s *MemoryStore) GetUserProfile(userID string) (domain.UserProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, ok := s.profiles[userID]
	if !ok {
		return domain.UserProfile{}, ErrNotFound
	}
	return *profile, nil
}

func (s *MemoryStore) SaveUserProfile(userID string, req domain.UpsertUserProfileRequest) (domain.UserProfile, error) {
	if req.Age < 14 || req.Age > 120 || strings.TrimSpace(req.Occupation) == "" || req.FreeHoursPerWeek <= 0 || req.FreeHoursPerWeek > 168 || req.AvailableBudget < 0 {
		return domain.UserProfile{}, ErrValidation
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	createdAt := now
	if current, ok := s.profiles[userID]; ok {
		createdAt = current.CreatedAt
	}
	profile := &domain.UserProfile{UserID: userID, Age: req.Age, Occupation: strings.TrimSpace(req.Occupation), FreeHoursPerWeek: req.FreeHoursPerWeek, AvailableBudget: req.AvailableBudget, Constraints: strings.TrimSpace(req.Constraints), CreatedAt: createdAt, UpdatedAt: now}
	s.profiles[userID] = profile
	return *profile, nil
}

func (s *MemoryStore) RegisterUser(req domain.RegisterRequest) (domain.AuthResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	email := normalizeEmail(req.Email)
	if email == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Name) == "" {
		return domain.AuthResponse{}, ErrValidation
	}

	if _, exists := s.usersByEmail[email]; exists {
		return domain.AuthResponse{}, ErrConflict
	}

	record := &domain.UserRecord{
		User: domain.User{
			ID:        newID(),
			Email:     email,
			Name:      strings.TrimSpace(req.Name),
			AvatarURL: nil,
			Provider:  domain.AuthProviderLocal,
			CreatedAt: time.Now().UTC(),
		},
		PasswordHash: hashPassword(req.Password),
	}

	s.users[record.ID] = record
	s.usersByEmail[email] = record.ID

	token := newToken()
	s.sessions[token] = record.ID

	return domain.AuthResponse{
		AccessToken: token,
		User:        record.User,
	}, nil
}

func (s *MemoryStore) Login(req domain.LoginRequest) (domain.AuthResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	email := normalizeEmail(req.Email)
	if email == "" || strings.TrimSpace(req.Password) == "" {
		return domain.AuthResponse{}, ErrValidation
	}

	userID, ok := s.usersByEmail[email]
	if !ok {
		return domain.AuthResponse{}, ErrInvalidAuth
	}

	record := s.users[userID]
	if !comparePassword(record.PasswordHash, req.Password) {
		return domain.AuthResponse{}, ErrInvalidAuth
	}

	token := newToken()
	s.sessions[token] = record.ID

	return domain.AuthResponse{
		AccessToken: token,
		User:        record.User,
	}, nil
}

func (s *MemoryStore) UserByToken(token string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.sessions[token]
	if !ok {
		return domain.User{}, ErrUnauthorized
	}

	record, ok := s.users[userID]
	if !ok {
		return domain.User{}, ErrUnauthorized
	}

	return record.User, nil
}

func (s *MemoryStore) GetAISettings(userID string) (domain.AISettingsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	settings, ok := s.aiSettings[userID]
	if !ok {
		return domain.AISettingsResponse{}, nil
	}
	return domain.AISettingsResponse{Provider: settings.Provider, MaskedKey: maskAPIKey(settings.APIKey)}, nil
}

func (s *MemoryStore) GetAISettingsSecret(userID string) (domain.AISettingsRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	settings, ok := s.aiSettings[userID]
	if !ok {
		return domain.AISettingsRequest{}, nil
	}
	return *settings, nil
}

func (s *MemoryStore) SaveAISettings(userID string, req domain.AISettingsRequest) (domain.AISettingsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider != string(domain.AIProviderGemini) && provider != string(domain.AIProviderDeepSeek) {
		return domain.AISettingsResponse{}, ErrValidation
	}
	if strings.TrimSpace(req.APIKey) == "" {
		return domain.AISettingsResponse{}, ErrValidation
	}

	s.aiSettings[userID] = &domain.AISettingsRequest{
		Provider: provider,
		APIKey:   strings.TrimSpace(req.APIKey),
	}
	return domain.AISettingsResponse{Provider: provider, MaskedKey: maskAPIKey(req.APIKey)}, nil
}

func (s *MemoryStore) SaveAnalyzedRequest(userID, originalText string, provider domain.AIProviderName, analysis domain.AnalyzeFreeformResponse) (domain.AnalyzeFreeformResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(userID) == "" || strings.TrimSpace(originalText) == "" {
		return domain.AnalyzeFreeformResponse{}, ErrValidation
	}

	now := time.Now().UTC()
	requestID := newID()
	tracker := analysis.ActivityTracker
	tracker.Stage = "questions_ready"

	record := &domain.PlanningRequest{
		ID:                  requestID,
		UserID:              userID,
		OriginalText:        strings.TrimSpace(originalText),
		Provider:            provider,
		InterpretedGoal:     analysis.InterpretedGoal,
		Feasibility:         analysis.Feasibility,
		ClarifyingQuestions: append([]domain.ClarifyingQuestion(nil), analysis.ClarifyingQuestions...),
		Answers:             []domain.RequestAnswer{},
		ActivityTracker:     tracker,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	s.planningRequests[requestID] = record

	analysis.RequestID = requestID
	analysis.OriginalText = strings.TrimSpace(originalText)
	analysis.ActivityTracker = tracker
	return analysis, nil
}

func (s *MemoryStore) SaveRequestAnswers(userID, requestID string, req domain.SubmitRequestAnswers) (domain.SubmitRequestAnswersResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.planningRequests[requestID]
	if !ok || record.UserID != userID {
		return domain.SubmitRequestAnswersResponse{}, ErrNotFound
	}

	if len(req.Answers) == 0 {
		return domain.SubmitRequestAnswersResponse{}, ErrValidation
	}

	questionsByID := make(map[string]domain.ClarifyingQuestion, len(record.ClarifyingQuestions))
	for _, question := range record.ClarifyingQuestions {
		questionsByID[question.ID] = question
	}

	answersByQuestion := make(map[string]domain.RequestAnswer, len(record.Answers)+len(req.Answers))
	for _, answer := range record.Answers {
		answersByQuestion[answer.QuestionID] = answer
	}
	for _, answer := range req.Answers {
		if strings.TrimSpace(answer.QuestionID) == "" || len(answer.Value) == 0 {
			return domain.SubmitRequestAnswersResponse{}, ErrValidation
		}
		question, ok := questionsByID[answer.QuestionID]
		if !ok || !validAnswerValue(question, answer.Value) {
			return domain.SubmitRequestAnswersResponse{}, ErrValidation
		}
		answersByQuestion[answer.QuestionID] = answer
	}

	record.Answers = make([]domain.RequestAnswer, 0, len(answersByQuestion))
	for _, question := range record.ClarifyingQuestions {
		if answer, ok := answersByQuestion[question.ID]; ok {
			record.Answers = append(record.Answers, answer)
		}
	}

	record.ActivityTracker.Stage = "answers_saved"
	record.ActivityTracker.Items = append(record.ActivityTracker.Items, domain.ActivitySignal{
		Label: "Answers saved",
		Value: fmt.Sprintf("%d", len(record.Answers)),
	})
	record.UpdatedAt = time.Now().UTC()

	return domain.SubmitRequestAnswersResponse{
		RequestID:       requestID,
		AnswersSaved:    len(record.Answers),
		ActivityTracker: record.ActivityTracker,
	}, nil
}

func (s *MemoryStore) GetPlanningRequest(userID, requestID string) (domain.PlanningRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.planningRequests[requestID]
	if !ok || record.UserID != userID {
		return domain.PlanningRequest{}, ErrNotFound
	}

	return *record, nil
}

func (s *MemoryStore) SaveRequestEvaluation(userID, requestID string, evaluation domain.EvaluateRequestResponse) (domain.EvaluateRequestResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.planningRequests[requestID]
	if !ok || record.UserID != userID {
		return domain.EvaluateRequestResponse{}, ErrNotFound
	}

	record.Feasibility = evaluation.Feasibility
	if evaluation.ClarifyingQuestions != nil {
		record.ClarifyingQuestions = mergeStoreQuestions(record.ClarifyingQuestions, evaluation.ClarifyingQuestions...)
	}
	record.ActivityTracker = evaluation.ActivityTracker
	record.ActivityTracker.Stage = "feasibility_evaluated"
	record.UpdatedAt = time.Now().UTC()

	evaluation.RequestID = requestID
	evaluation.ClarifyingQuestions = record.ClarifyingQuestions
	evaluation.ActivityTracker = record.ActivityTracker
	return evaluation, nil
}

func (s *MemoryStore) CreateGeneratedPlan(userID, requestID string, plan domain.GeneratedPlan) (domain.GeneratePlanResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.planningRequests[requestID]
	if !ok || record.UserID != userID {
		return domain.GeneratePlanResponse{}, ErrNotFound
	}

	if strings.TrimSpace(plan.Goal.Title) == "" || len(plan.Tasks) == 0 {
		return domain.GeneratePlanResponse{}, ErrValidation
	}
	if !record.Feasibility.CanGeneratePlan {
		return domain.GeneratePlanResponse{}, ErrInvalidState
	}

	now := time.Now().UTC()
	goal := &domain.GoalRecord{
		Goal: domain.Goal{
			ID:                    newID(),
			Title:                 strings.TrimSpace(plan.Goal.Title),
			Description:           strings.TrimSpace(plan.Goal.Description),
			Category:              strings.TrimSpace(plan.Goal.Category),
			Priority:              normalizePriority(plan.Goal.Priority),
			Status:                domain.GoalStatusActive,
			Deadline:              normalizeDatePointer(plan.Goal.Deadline),
			AvailableHoursPerWeek: plan.Goal.AvailableHoursPerWeek,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		UserID: userID,
	}
	if goal.Category == "" {
		goal.Category = "Personal Development"
	}
	if goal.Description == "" {
		goal.Description = record.OriginalText
	}
	if goal.AvailableHoursPerWeek <= 0 {
		goal.AvailableHoursPerWeek = 10
	}

	if s.findUserGoalTitleConflictLocked(userID, goal.Title, "") {
		goal.Title = fmt.Sprintf("%s (%s)", goal.Title, now.Format("2006-01-02 15:04"))
	}

	s.goals[goal.ID] = goal
	s.goalOrder = append(s.goalOrder, goal.ID)

	context := domain.GoalContext{
		ID:              newID(),
		GoalID:          goal.ID,
		CurrentLevel:    strings.TrimSpace(plan.Context.CurrentLevel),
		ConstraintsText: strings.TrimSpace(plan.Context.ConstraintsText),
		Preferences:     append([]string(nil), plan.Context.Preferences...),
		ExpectedResult:  strings.TrimSpace(plan.Context.ExpectedResult),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.goalContext[goal.ID] = &context

	orderIndex := 1
	taskIDsByTitle := make(map[string]string)
	pendingDependencies := make(map[string][]string)
	for _, generatedTask := range plan.Tasks {
		s.createGeneratedTaskLocked(goal.ID, nil, generatedTask, &orderIndex, now, taskIDsByTitle, pendingDependencies)
	}
	for taskID, titles := range pendingDependencies {
		for _, title := range titles {
			dependsOnID, ok := taskIDsByTitle[strings.TrimSpace(title)]
			if !ok || dependsOnID == taskID {
				continue
			}
			dep := &domain.TaskDependency{ID: newID(), TaskID: taskID, DependsOnTaskID: dependsOnID, CreatedAt: now}
			s.deps[dep.ID] = dep
			s.depsByTask[taskID] = append(s.depsByTask[taskID], dep.ID)
		}
	}

	generation := domain.GenerationDetails{
		Generation: domain.Generation{
			ID:        newID(),
			GoalID:    goal.ID,
			Status:    domain.GenerationStatusSuccess,
			ModelName: "mock-ai-provider",
			CreatedAt: now,
		},
		Prompt:        record.OriginalText,
		RawAIResponse: "Generated from structured request flow.",
	}
	s.generations[generation.ID] = &generation
	s.generationsByGoal[goal.ID] = append(s.generationsByGoal[goal.ID], generation.ID)

	version := domain.PlanVersion{
		ID:            newID(),
		GoalID:        goal.ID,
		GenerationID:  generation.ID,
		VersionNumber: 1,
		IsActive:      true,
		CreatedAt:     now,
	}
	s.versions[version.ID] = &version
	s.versionsByGoal[goal.ID] = append(s.versionsByGoal[goal.ID], version.ID)

	record.ActivityTracker.Stage = "plan_generated"
	record.ActivityTracker.Items = append(record.ActivityTracker.Items, domain.ActivitySignal{
		Label: "Generated goal",
		Value: goal.Title,
	})
	record.UpdatedAt = now

	progress, _ := s.getProgressLocked(goal.ID)
	return domain.GeneratePlanResponse{
		RequestID:   requestID,
		Feasibility: record.Feasibility,
		Goal:        goal.Goal,
		Tasks:       buildTaskTree(s.tasksForGoalLocked(goal.ID)),
		Progress:    progress,
	}, nil
}

func (s *MemoryStore) ReplaceGoalPlan(userID, goalID string, plan domain.GeneratedPlan) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	goal, ok := s.goals[goalID]
	if !ok || goal.UserID != userID || len(plan.Tasks) == 0 {
		return 0, ErrNotFound
	}
	for _, task := range s.tasksForGoalLocked(goalID) {
		s.deleteTaskCascadeLocked(task.ID)
	}
	goal.Title = strings.TrimSpace(plan.Goal.Title)
	if goal.Title == "" {
		goal.Title = "Цель"
	}
	goal.Description = strings.TrimSpace(plan.Goal.Description)
	goal.Deadline = normalizeDatePointer(plan.Goal.Deadline)
	goal.Status = domain.GoalStatusActive
	goal.UpdatedAt = time.Now().UTC()
	order := 1
	ids := map[string]string{}
	pending := map[string][]string{}
	for _, generated := range plan.Tasks {
		s.createGeneratedTaskLocked(goalID, nil, generated, &order, goal.UpdatedAt, ids, pending)
	}
	for taskID, titles := range pending {
		for _, title := range titles {
			if dependsOnID, ok := ids[strings.TrimSpace(title)]; ok && dependsOnID != taskID {
				dep := &domain.TaskDependency{ID: newID(), TaskID: taskID, DependsOnTaskID: dependsOnID, CreatedAt: goal.UpdatedAt}
				s.deps[dep.ID] = dep
				s.depsByTask[taskID] = append(s.depsByTask[taskID], dep.ID)
			}
		}
	}
	return order - 1, nil
}

func (s *MemoryStore) CreateOAuthUserSession(email, name, avatarURL string, provider domain.AuthProvider) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if email == "" {
		return "", ErrValidation
	}
	if name == "" {
		name = email
	}

	var record *domain.UserRecord

	if userID, exists := s.usersByEmail[email]; exists {
		record = s.users[userID]
	} else {
		var avatar *string
		if strings.TrimSpace(avatarURL) != "" {
			trimmedAvatar := strings.TrimSpace(avatarURL)
			avatar = &trimmedAvatar
		}
		record = &domain.UserRecord{
			User: domain.User{
				ID:        newID(),
				Email:     email,
				Name:      name,
				AvatarURL: avatar,
				Provider:  provider,
				CreatedAt: time.Now().UTC(),
			},
		}
		s.users[record.ID] = record
		s.usersByEmail[email] = record.ID
	}

	token := newToken()
	s.sessions[token] = record.ID
	return token, nil
}

func (s *MemoryStore) ListGoals(userID string) []domain.Goal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Goal, 0, len(s.goalOrder))
	for _, goalID := range s.goalOrder {
		goal, ok := s.goals[goalID]
		if !ok || goal.UserID != userID {
			continue
		}
		result = append(result, goal.Goal)
	}

	slices.SortFunc(result, func(a, b domain.Goal) int {
		return b.UpdatedAt.Compare(a.UpdatedAt)
	})

	return result
}

func (s *MemoryStore) CreateGoal(userID string, req domain.CreateGoalRequest) (domain.Goal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Description) == "" || strings.TrimSpace(req.Category) == "" {
		return domain.Goal{}, ErrValidation
	}

	now := time.Now().UTC()
	record := &domain.GoalRecord{
		Goal: domain.Goal{
			ID:                    newID(),
			Title:                 strings.TrimSpace(req.Title),
			Description:           strings.TrimSpace(req.Description),
			Category:              strings.TrimSpace(req.Category),
			Priority:              normalizePriority(req.Priority),
			Status:                domain.GoalStatusDraft,
			Deadline:              normalizeDatePointer(req.Deadline),
			AvailableHoursPerWeek: req.AvailableHoursPerWeek,
			CreatedAt:             now,
			UpdatedAt:             now,
		},
		UserID: userID,
	}

	if conflict := s.findUserGoalTitleConflictLocked(userID, record.Title, ""); conflict {
		return domain.Goal{}, ErrConflict
	}

	s.goals[record.ID] = record
	s.goalOrder = append(s.goalOrder, record.ID)
	s.ensureBaselineVersionLocked(record.ID)
	return record.Goal, nil
}

func (s *MemoryStore) GetGoal(userID, goalID string) (domain.Goal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.goals[goalID]
	if !ok || record.UserID != userID {
		return domain.Goal{}, ErrNotFound
	}

	return record.Goal, nil
}

func (s *MemoryStore) UpdateGoal(userID, goalID string, req domain.UpdateGoalRequest) (domain.Goal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.goals[goalID]
	if !ok || record.UserID != userID {
		return domain.Goal{}, ErrNotFound
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return domain.Goal{}, ErrValidation
		}
		if s.findUserGoalTitleConflictLocked(userID, title, goalID) {
			return domain.Goal{}, ErrConflict
		}
		record.Title = title
	}
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description == "" {
			return domain.Goal{}, ErrValidation
		}
		record.Description = description
	}

	if req.Category != nil {
		category := strings.TrimSpace(*req.Category)
		if category == "" {
			return domain.Goal{}, ErrValidation
		}
		record.Category = category
	}

	if req.Priority != nil {
		record.Priority = normalizePriority(*req.Priority)
	}

	if req.Status != nil {
		record.Status = normalizeGoalStatus(*req.Status)
	}

	if req.Deadline != nil {
		record.Deadline = normalizeDatePointer(req.Deadline)
	}

	if req.AvailableHoursPerWeek != nil {
		record.AvailableHoursPerWeek = *req.AvailableHoursPerWeek
	}

	record.UpdatedAt = time.Now().UTC()
	return record.Goal, nil
}

func (s *MemoryStore) DeleteGoal(userID, goalID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.goals[goalID]
	if !ok || record.UserID != userID {
		return ErrNotFound
	}

	delete(s.goals, goalID)
	s.goalOrder = removeString(s.goalOrder, goalID)
	delete(s.goalContext, goalID)

	for taskID, task := range s.tasks {
		if task.GoalID != goalID {
			continue
		}
		delete(s.tasks, taskID)
		s.taskOrder = removeString(s.taskOrder, taskID)
		delete(s.depsByTask, taskID)
	}

	for depID, dep := range s.deps {
		if task, ok := s.tasks[dep.TaskID]; !ok || task.GoalID == goalID {
			delete(s.deps, depID)
		}
	}

	for _, versionID := range s.versionsByGoal[goalID] {
		delete(s.versions, versionID)
	}
	delete(s.versionsByGoal, goalID)

	for _, generationID := range s.generationsByGoal[goalID] {
		delete(s.generations, generationID)
		for _, feedbackID := range s.feedbackByGen[generationID] {
			delete(s.feedback, feedbackID)
		}
		delete(s.feedbackByGen, generationID)
	}
	delete(s.generationsByGoal, goalID)

	return nil
}

func (s *MemoryStore) GetGoalContext(userID, goalID string) (domain.GoalContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.GoalContext{}, ErrNotFound
	}

	context, ok := s.goalContext[goalID]
	if !ok {
		return domain.GoalContext{}, ErrNotFound
	}

	return *context, nil
}

func (s *MemoryStore) CreateGoalContext(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.GoalContext{}, ErrNotFound
	}

	if _, exists := s.goalContext[goalID]; exists {
		return domain.GoalContext{}, ErrConflict
	}

	context := domain.GoalContext{
		ID:              newID(),
		GoalID:          goalID,
		CurrentLevel:    strings.TrimSpace(req.CurrentLevel),
		ConstraintsText: strings.TrimSpace(req.ConstraintsText),
		Preferences:     append([]string(nil), req.Preferences...),
		ExpectedResult:  strings.TrimSpace(req.ExpectedResult),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	s.goalContext[goalID] = &context
	return context, nil
}

func (s *MemoryStore) UpdateGoalContext(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.GoalContext{}, ErrNotFound
	}

	context, exists := s.goalContext[goalID]
	if !exists {
		return domain.GoalContext{}, ErrNotFound
	}

	context.CurrentLevel = strings.TrimSpace(req.CurrentLevel)
	context.ConstraintsText = strings.TrimSpace(req.ConstraintsText)
	context.Preferences = append([]string(nil), req.Preferences...)
	context.ExpectedResult = strings.TrimSpace(req.ExpectedResult)
	context.UpdatedAt = time.Now().UTC()

	return *context, nil
}

func (s *MemoryStore) ListTasksByGoal(userID, goalID string) ([]domain.TaskTreeNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return nil, ErrNotFound
	}

	flat := s.tasksForGoalLocked(goalID)
	return buildTaskTree(flat), nil
}

func (s *MemoryStore) CreateTask(userID, goalID string, req domain.CreateTaskRequest) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.Task{}, ErrNotFound
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return domain.Task{}, ErrValidation
	}

	if req.ParentTaskID != nil {
		parent, ok := s.tasks[*req.ParentTaskID]
		if !ok || parent.GoalID != goalID {
			return domain.Task{}, ErrNotFound
		}
	}

	orderIndex := nextOrderIndex(s.tasksForGoalLocked(goalID))
	if req.OrderIndex != nil {
		orderIndex = *req.OrderIndex
	}

	status := domain.TaskStatusTodo
	if req.Status != nil {
		status = normalizeTaskStatus(*req.Status)
	}

	task := domain.Task{
		ID:             newID(),
		GoalID:         goalID,
		ParentTaskID:   req.ParentTaskID,
		Title:          title,
		Description:    strings.TrimSpace(req.Description),
		Status:         status,
		Priority:       normalizePriority(req.Priority),
		EstimatedHours: req.EstimatedHours,
		Deadline:       normalizeDatePointer(req.Deadline),
		OrderIndex:     orderIndex,
		Source:         domain.TaskSourceManual,
		EditedByUser:   true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	s.tasks[task.ID] = &task
	s.taskOrder = append(s.taskOrder, task.ID)
	s.touchGoalLocked(goalID)
	return task, nil
}

func (s *MemoryStore) GetTask(userID, taskID string) (domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return domain.Task{}, ErrNotFound
	}

	if !s.userOwnsGoalLocked(userID, task.GoalID) {
		return domain.Task{}, ErrNotFound
	}

	return *task, nil
}

func (s *MemoryStore) UpdateTask(userID, taskID string, req domain.UpdateTaskRequest) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok || !s.userOwnsGoalLocked(userID, task.GoalID) {
		return domain.Task{}, ErrNotFound
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return domain.Task{}, ErrValidation
		}
		task.Title = title
	}
	if req.ParentTaskID != nil {
		parent, ok := s.tasks[*req.ParentTaskID]
		if !ok || parent.GoalID != task.GoalID || parent.ParentTaskID != nil || parent.ID == task.ID {
			return domain.Task{}, ErrValidation
		}
		task.ParentTaskID = req.ParentTaskID
	}

	if req.Description != nil {
		task.Description = strings.TrimSpace(*req.Description)
	}

	if req.Status != nil {
		task.Status = normalizeTaskStatus(*req.Status)
	}

	if req.Priority != nil {
		task.Priority = normalizePriority(*req.Priority)
	}

	if req.EstimatedHours != nil {
		task.EstimatedHours = *req.EstimatedHours
	}

	if req.Deadline != nil {
		task.Deadline = normalizeDatePointer(req.Deadline)
	}

	if req.OrderIndex != nil {
		task.OrderIndex = *req.OrderIndex
	}

	if req.EditedByUser != nil {
		task.EditedByUser = *req.EditedByUser
	} else {
		task.EditedByUser = true
	}

	task.UpdatedAt = time.Now().UTC()
	s.touchGoalLocked(task.GoalID)
	return *task, nil
}

func (s *MemoryStore) DeleteTask(userID, taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok || !s.userOwnsGoalLocked(userID, task.GoalID) {
		return ErrNotFound
	}

	s.deleteTaskCascadeLocked(taskID)
	s.deletedTasksByGoal[task.GoalID]++
	s.touchGoalLocked(task.GoalID)
	return nil
}

func (s *MemoryStore) UpdateTaskStatus(userID, taskID string, status domain.TaskStatus) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok || !s.userOwnsGoalLocked(userID, task.GoalID) {
		return domain.Task{}, ErrNotFound
	}

	status = normalizeTaskStatus(status)
	if status == domain.TaskStatusDone {
		for _, depID := range s.depsByTask[taskID] {
			dep := s.deps[depID]
			if dependency := s.tasks[dep.DependsOnTaskID]; dependency != nil && dependency.Status != domain.TaskStatusDone {
				return domain.Task{}, ErrInvalidState
			}
		}
	}
	task.Status = status
	task.EditedByUser = true
	task.UpdatedAt = time.Now().UTC()
	s.touchGoalLocked(task.GoalID)
	return *task, nil
}

func (s *MemoryStore) ListDependencies(userID, taskID string) ([]domain.TaskDependency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok || !s.userOwnsGoalLocked(userID, task.GoalID) {
		return nil, ErrNotFound
	}

	depIDs := s.depsByTask[taskID]
	result := make([]domain.TaskDependency, 0, len(depIDs))
	for _, depID := range depIDs {
		if dep, ok := s.deps[depID]; ok {
			result = append(result, *dep)
		}
	}

	slices.SortFunc(result, func(a, b domain.TaskDependency) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	return result, nil
}

func (s *MemoryStore) CreateDependency(userID, taskID string, req domain.CreateTaskDependencyRequest) (domain.TaskDependency, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok || !s.userOwnsGoalLocked(userID, task.GoalID) {
		return domain.TaskDependency{}, ErrNotFound
	}

	dependsOn, ok := s.tasks[req.DependsOnTaskID]
	if !ok || dependsOn.GoalID != task.GoalID {
		return domain.TaskDependency{}, ErrNotFound
	}

	for _, depID := range s.depsByTask[taskID] {
		if dep, ok := s.deps[depID]; ok && dep.DependsOnTaskID == req.DependsOnTaskID {
			return domain.TaskDependency{}, ErrConflict
		}
	}

	dependency := domain.TaskDependency{
		ID:              newID(),
		TaskID:          taskID,
		DependsOnTaskID: req.DependsOnTaskID,
		CreatedAt:       time.Now().UTC(),
	}

	s.deps[dependency.ID] = &dependency
	s.depsByTask[taskID] = append(s.depsByTask[taskID], dependency.ID)
	return dependency, nil
}

func (s *MemoryStore) DeleteDependency(userID, taskID, dependencyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok || !s.userOwnsGoalLocked(userID, task.GoalID) {
		return ErrNotFound
	}

	dependency, ok := s.deps[dependencyID]
	if !ok || dependency.TaskID != taskID {
		return ErrNotFound
	}

	delete(s.deps, dependencyID)
	s.depsByTask[taskID] = removeString(s.depsByTask[taskID], dependencyID)
	return nil
}

func (s *MemoryStore) GetProgress(userID, goalID string) (domain.ProgressResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.ProgressResponse{}, ErrNotFound
	}

	tasks := s.tasksForGoalLocked(goalID)
	response := domain.ProgressResponse{GoalID: goalID}
	response.TotalTasks = len(tasks)

	for _, task := range tasks {
		response.TotalEstimatedHours += task.EstimatedHours
		if task.Status == domain.TaskStatusDone {
			response.CompletedTasks++
			response.CompletedEstimatedHours += task.EstimatedHours
		}
	}

	if response.TotalTasks > 0 {
		response.SimpleProgressPercent = round2(float64(response.CompletedTasks) / float64(response.TotalTasks) * 100)
	}

	if response.TotalEstimatedHours > 0 {
		response.WeightedProgressPercent = round2(response.CompletedEstimatedHours / response.TotalEstimatedHours * 100)
	}

	return response, nil
}

func (s *MemoryStore) GetFeasibility(userID, goalID string) (domain.FeasibilityResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	goal, ok := s.goals[goalID]
	if !ok || goal.UserID != userID {
		return domain.FeasibilityResponse{}, ErrNotFound
	}

	response := domain.FeasibilityResponse{GoalID: goalID}
	tasks := s.tasksForGoalLocked(goalID)
	for _, task := range tasks {
		response.RequiredHours += task.EstimatedHours
	}

	if goal.Deadline == nil {
		response.Status = domain.FeasibilityStatusUnknown
		response.Message = "Set a deadline to calculate planning feasibility."
		return response, nil
	}

	deadlineTime, err := time.Parse("2006-01-02", *goal.Deadline)
	if err != nil {
		response.Status = domain.FeasibilityStatusUnknown
		response.Message = "The goal deadline has an invalid format."
		return response, nil
	}

	hoursUntilDeadline := deadlineTime.Sub(time.Now().UTC()).Hours()
	if hoursUntilDeadline < 0 {
		hoursUntilDeadline = 0
	}

	response.WeeksUntilDeadline = int(math.Ceil(hoursUntilDeadline / (24 * 7)))
	if response.WeeksUntilDeadline == 0 && hoursUntilDeadline > 0 {
		response.WeeksUntilDeadline = 1
	}

	response.AvailableHours = round2(float64(response.WeeksUntilDeadline) * goal.AvailableHoursPerWeek)

	switch {
	case response.RequiredHours == 0:
		response.Status = domain.FeasibilityStatusUnknown
		response.Message = "Add estimated hours to tasks to evaluate feasibility."
	case response.AvailableHours >= response.RequiredHours:
		response.Status = domain.FeasibilityStatusRealistic
		response.Message = "The current pace looks realistic for the selected deadline."
	case response.AvailableHours*1.2 >= response.RequiredHours:
		response.Status = domain.FeasibilityStatusRisky
		response.Message = "The plan is tight. Consider reducing scope or adding more weekly hours."
	default:
		response.Status = domain.FeasibilityStatusUnrealistic
		response.Message = "At the current pace, the deadline is likely unrealistic."
	}

	return response, nil
}

func (s *MemoryStore) GetQuality(userID, goalID string) (domain.QualityResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.QualityResponse{}, ErrNotFound
	}

	response := domain.QualityResponse{GoalID: goalID}
	score := 35.0

	if context, ok := s.goalContext[goalID]; ok {
		score += 20
		if len(context.Preferences) > 0 {
			score += 5
			response.Details = append(response.Details, "Context preferences are defined.")
		}
		response.Details = append(response.Details, "Goal context is documented.")
	}

	tasks := s.tasksForGoalLocked(goalID)
	if len(tasks) >= 3 {
		score += 20
		response.Details = append(response.Details, "The plan has at least three task nodes.")
	} else if len(tasks) > 0 {
		score += 10
		response.Details = append(response.Details, "The plan has an initial task structure.")
	}

	hasDeadlines := false
	hasPriorities := false
	for _, task := range tasks {
		if task.Deadline != nil {
			hasDeadlines = true
		}
		if task.Priority != "" {
			hasPriorities = true
		}
	}

	if hasDeadlines {
		score += 10
		response.Details = append(response.Details, "Tasks include deadline guidance.")
	}

	if hasPriorities {
		score += 10
		response.Details = append(response.Details, "Tasks are prioritized.")
	}

	if len(s.depsByTask) > 0 {
		score += 10
		response.Details = append(response.Details, "Dependency relationships are captured.")
	}

	response.Score = math.Min(round2(score), 100)
	if len(response.Details) == 0 {
		response.Details = []string{"Add context and structured tasks to improve plan quality."}
	}

	return response, nil
}

func (s *MemoryStore) GetMetrics(userID, goalID string) (domain.MetricsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.MetricsResponse{}, ErrNotFound
	}

	response := domain.MetricsResponse{GoalID: goalID}
	tasks := s.tasksForGoalLocked(goalID)
	totalTasks := len(tasks)

	for _, task := range tasks {
		if task.Source == domain.TaskSourceAI {
			response.TasksCreatedByAI++
		}
		if task.EditedByUser {
			response.TasksEditedByUser++
		}
		if task.Status == domain.TaskStatusDone {
			response.TasksCompleted++
		}
	}

	response.TasksDeletedByUser = s.deletedTasksByGoal[goalID]
	if generationIDs := s.generationsByGoal[goalID]; len(generationIDs) > 1 {
		response.RegenerationCount = len(generationIDs) - 1
	}

	if totalTasks > 0 {
		response.AcceptedTasksPercent = round2(float64(response.TasksCompleted) / float64(totalTasks) * 100)
	}

	return response, nil
}

func (s *MemoryStore) ListPlanVersions(userID, goalID string) ([]domain.PlanVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return nil, ErrNotFound
	}

	versionIDs := s.versionsByGoal[goalID]
	result := make([]domain.PlanVersion, 0, len(versionIDs))
	for _, versionID := range versionIDs {
		if version, ok := s.versions[versionID]; ok {
			result = append(result, *version)
		}
	}

	slices.SortFunc(result, func(a, b domain.PlanVersion) int {
		return a.VersionNumber - b.VersionNumber
	})
	return result, nil
}

func (s *MemoryStore) GetPlanVersion(userID, goalID, versionID string) (domain.PlanVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.PlanVersion{}, ErrNotFound
	}

	version, ok := s.versions[versionID]
	if !ok || version.GoalID != goalID {
		return domain.PlanVersion{}, ErrNotFound
	}

	return *version, nil
}

func (s *MemoryStore) ActivatePlanVersion(userID, goalID, versionID string) (domain.PlanVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return domain.PlanVersion{}, ErrNotFound
	}

	version, ok := s.versions[versionID]
	if !ok || version.GoalID != goalID {
		return domain.PlanVersion{}, ErrNotFound
	}

	for _, currentID := range s.versionsByGoal[goalID] {
		if current, ok := s.versions[currentID]; ok {
			current.IsActive = current.ID == versionID
		}
	}

	return *version, nil
}

func (s *MemoryStore) ListGenerations(userID, goalID string) ([]domain.Generation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.userOwnsGoalLocked(userID, goalID) {
		return nil, ErrNotFound
	}

	generationIDs := s.generationsByGoal[goalID]
	result := make([]domain.Generation, 0, len(generationIDs))
	for _, generationID := range generationIDs {
		if generation, ok := s.generations[generationID]; ok {
			result = append(result, generation.Generation)
		}
	}

	slices.SortFunc(result, func(a, b domain.Generation) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})
	return result, nil
}

func (s *MemoryStore) GetGeneration(userID, generationID string) (domain.GenerationDetails, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	generation, ok := s.generations[generationID]
	if !ok {
		return domain.GenerationDetails{}, ErrNotFound
	}

	if !s.userOwnsGoalLocked(userID, generation.GoalID) {
		return domain.GenerationDetails{}, ErrNotFound
	}

	return *generation, nil
}

func (s *MemoryStore) CreateFeedback(userID, generationID string, req domain.FeedbackRequest) (domain.Feedback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	generation, ok := s.generations[generationID]
	if !ok || !s.userOwnsGoalLocked(userID, generation.GoalID) {
		return domain.Feedback{}, ErrNotFound
	}

	if req.Rating < 1 || req.Rating > 5 {
		return domain.Feedback{}, ErrValidation
	}

	feedback := domain.Feedback{
		ID:           newID(),
		GenerationID: generationID,
		Rating:       req.Rating,
		Comment:      strings.TrimSpace(req.Comment),
		CreatedAt:    time.Now().UTC(),
	}

	s.feedback[feedback.ID] = &feedback
	s.feedbackByGen[generationID] = append(s.feedbackByGen[generationID], feedback.ID)
	return feedback, nil
}

func (s *MemoryStore) DecomposeGoal(userID, goalID string, req domain.DecomposeGoalRequest, mode string) (domain.DecomposeGoalResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	goal, ok := s.goals[goalID]
	if !ok || goal.UserID != userID {
		return domain.DecomposeGoalResponse{}, ErrNotFound
	}

	baseTitle := goal.Title
	if baseTitle == "" {
		baseTitle = "Goal plan"
	}

	detailLevel := string(req.DetailLevel)
	if detailLevel == "" {
		detailLevel = string(domain.DetailLevelHigh)
	}

	now := time.Now().UTC()
	generation := domain.GenerationDetails{
		Generation: domain.Generation{
			ID:        newID(),
			GoalID:    goalID,
			Status:    domain.GenerationStatusSuccess,
			ModelName: "goalmind-synth-v1",
			CreatedAt: now,
		},
		Prompt:        fmt.Sprintf("Generate a %s detail plan for goal %q", strings.ToLower(detailLevel), baseTitle),
		RawAIResponse: fmt.Sprintf("Synthetic %s decomposition generated at %s.", mode, now.Format(time.RFC3339)),
	}

	s.generations[generation.ID] = &generation
	s.generationsByGoal[goalID] = append(s.generationsByGoal[goalID], generation.ID)

	version := domain.PlanVersion{
		ID:            newID(),
		GoalID:        goalID,
		GenerationID:  generation.ID,
		VersionNumber: len(s.versionsByGoal[goalID]) + 1,
		IsActive:      true,
		CreatedAt:     now,
	}
	for _, existingID := range s.versionsByGoal[goalID] {
		if existing, ok := s.versions[existingID]; ok {
			existing.IsActive = false
		}
	}
	s.versions[version.ID] = &version
	s.versionsByGoal[goalID] = append(s.versionsByGoal[goalID], version.ID)

	tasksCreated := 0
	stageNames := []string{"Research", "Build", "Validate"}
	for index, stageName := range stageNames {
		task := domain.Task{
			ID:             newID(),
			GoalID:         goalID,
			Title:          fmt.Sprintf("%s: %s", stageName, baseTitle),
			Description:    fmt.Sprintf("AI-generated %s stage for %s.", strings.ToLower(stageName), baseTitle),
			Status:         domain.TaskStatusTodo,
			Priority:       domain.GoalPriorityHigh,
			EstimatedHours: float64(4 + index*2),
			Deadline:       goal.Deadline,
			OrderIndex:     nextOrderIndex(s.tasksForGoalLocked(goalID)),
			Source:         domain.TaskSourceAI,
			EditedByUser:   false,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		s.tasks[task.ID] = &task
		s.taskOrder = append(s.taskOrder, task.ID)
		tasksCreated++
	}

	goal.Status = domain.GoalStatusActive
	goal.UpdatedAt = now

	quality, _ := s.getQualityLocked(goalID)
	feasibility, _ := s.getFeasibilityLocked(goalID)

	return domain.DecomposeGoalResponse{
		GoalID:            goalID,
		GenerationID:      generation.ID,
		PlanVersionID:     version.ID,
		TasksCreated:      tasksCreated,
		QualityScore:      quality.Score,
		FeasibilityStatus: feasibility.Status,
	}, nil
}

func (s *MemoryStore) DecomposeTask(userID, taskID string, req domain.DecomposeGoalRequest) (domain.DecomposeGoalResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok || !s.userOwnsGoalLocked(userID, task.GoalID) {
		return domain.DecomposeGoalResponse{}, ErrNotFound
	}

	now := time.Now().UTC()
	activeVersionID := s.ensureBaselineVersionLocked(task.GoalID)
	if activeVersion, ok := s.versions[activeVersionID]; ok {
		for _, versionID := range s.versionsByGoal[task.GoalID] {
			if version, exists := s.versions[versionID]; exists {
				version.IsActive = version.ID == activeVersion.ID
			}
		}
	}

	generation := domain.GenerationDetails{
		Generation: domain.Generation{
			ID:        newID(),
			GoalID:    task.GoalID,
			Status:    domain.GenerationStatusSuccess,
			ModelName: "goalmind-synth-v1",
			CreatedAt: now,
		},
		Prompt:        fmt.Sprintf("Expand task %q", task.Title),
		RawAIResponse: fmt.Sprintf("Expanded task %q into focused subtasks.", task.Title),
	}
	s.generations[generation.ID] = &generation
	s.generationsByGoal[task.GoalID] = append(s.generationsByGoal[task.GoalID], generation.ID)

	tasksCreated := 0
	childTitles := []string{"Outline the work", "Deliver the result"}
	for index, childTitle := range childTitles {
		parentID := task.ID
		child := domain.Task{
			ID:             newID(),
			GoalID:         task.GoalID,
			ParentTaskID:   &parentID,
			Title:          childTitle,
			Description:    fmt.Sprintf("AI-generated child task for %s.", task.Title),
			Status:         domain.TaskStatusTodo,
			Priority:       task.Priority,
			EstimatedHours: math.Max(1, task.EstimatedHours/2),
			Deadline:       task.Deadline,
			OrderIndex:     task.OrderIndex + index + 1,
			Source:         domain.TaskSourceAI,
			EditedByUser:   false,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		s.tasks[child.ID] = &child
		s.taskOrder = append(s.taskOrder, child.ID)
		tasksCreated++
	}

	s.touchGoalLocked(task.GoalID)
	quality, _ := s.getQualityLocked(task.GoalID)
	feasibility, _ := s.getFeasibilityLocked(task.GoalID)

	return domain.DecomposeGoalResponse{
		GoalID:            task.GoalID,
		GenerationID:      generation.ID,
		PlanVersionID:     activeVersionID,
		TasksCreated:      tasksCreated,
		QualityScore:      quality.Score,
		FeasibilityStatus: feasibility.Status,
	}, nil
}

func (s *MemoryStore) seed() {
	userID := s.seedUser("Anna Petrova", "anna@example.com", "strong-password-123", domain.AuthProviderLocal)
	s.seedGoalBundle(userID, "Launch GoalMind diploma MVP", "Prepare a polished demo version of GoalMind for thesis defense.")
	secondGoalID := s.seedGoalBundle(userID, "Build the Go backend", "Implement the first backend slice that supports the frontend without mocks.")

	if goal, ok := s.goals[secondGoalID]; ok {
		goal.Status = domain.GoalStatusDraft
	}
}

func (s *MemoryStore) seedUser(name, email, password string, provider domain.AuthProvider) string {
	record := &domain.UserRecord{
		User: domain.User{
			ID:        newID(),
			Email:     normalizeEmail(email),
			Name:      name,
			AvatarURL: nil,
			Provider:  provider,
			CreatedAt: time.Now().UTC().Add(-48 * time.Hour),
		},
		PasswordHash: hashPassword(password),
	}
	s.users[record.ID] = record
	s.usersByEmail[record.Email] = record.ID
	return record.ID
}

func (s *MemoryStore) seedGoalBundle(userID, title, description string) string {
	now := time.Now().UTC()
	deadline := now.AddDate(0, 1, 14).Format("2006-01-02")
	goal := &domain.GoalRecord{
		Goal: domain.Goal{
			ID:                    newID(),
			Title:                 title,
			Description:           description,
			Category:              "Education Product",
			Priority:              domain.GoalPriorityHigh,
			Status:                domain.GoalStatusActive,
			Deadline:              &deadline,
			AvailableHoursPerWeek: 16,
			CreatedAt:             now.Add(-24 * time.Hour),
			UpdatedAt:             now.Add(-6 * time.Hour),
		},
		UserID: userID,
	}
	s.goals[goal.ID] = goal
	s.goalOrder = append(s.goalOrder, goal.ID)

	context := &domain.GoalContext{
		ID:              newID(),
		GoalID:          goal.ID,
		CurrentLevel:    "Intermediate frontend implementation completed",
		ConstraintsText: "Work part-time, keep the scope thesis-ready, and prefer clear documentation.",
		Preferences:     []string{"Readable code", "Visible progress metrics", "Clean architecture later"},
		ExpectedResult:  "A demo-ready product with a working frontend and a progressive backend MVP.",
		CreatedAt:       now.Add(-20 * time.Hour),
		UpdatedAt:       now.Add(-4 * time.Hour),
	}
	s.goalContext[goal.ID] = context

	rootTask := domain.Task{
		ID:             newID(),
		GoalID:         goal.ID,
		Title:          "Define the backend scope",
		Description:    "Translate the frontend contract into a first server slice.",
		Status:         domain.TaskStatusDone,
		Priority:       domain.GoalPriorityHigh,
		EstimatedHours: 4,
		Deadline:       goal.Deadline,
		OrderIndex:     1,
		Source:         domain.TaskSourceManual,
		EditedByUser:   true,
		CreatedAt:      now.Add(-18 * time.Hour),
		UpdatedAt:      now.Add(-17 * time.Hour),
	}
	childParentID := rootTask.ID
	childTask := domain.Task{
		ID:             newID(),
		GoalID:         goal.ID,
		ParentTaskID:   &childParentID,
		Title:          "Implement auth and goals",
		Description:    "Ship the first real handlers that the dashboard can call.",
		Status:         domain.TaskStatusInProgress,
		Priority:       domain.GoalPriorityHigh,
		EstimatedHours: 10,
		Deadline:       goal.Deadline,
		OrderIndex:     2,
		Source:         domain.TaskSourceAI,
		EditedByUser:   false,
		CreatedAt:      now.Add(-16 * time.Hour),
		UpdatedAt:      now.Add(-2 * time.Hour),
	}
	finalTask := domain.Task{
		ID:             newID(),
		GoalID:         goal.ID,
		Title:          "Validate the frontend flow",
		Description:    "Turn off mocks and verify the main pages against the backend.",
		Status:         domain.TaskStatusTodo,
		Priority:       domain.GoalPriorityMedium,
		EstimatedHours: 6,
		Deadline:       goal.Deadline,
		OrderIndex:     3,
		Source:         domain.TaskSourceManual,
		EditedByUser:   true,
		CreatedAt:      now.Add(-12 * time.Hour),
		UpdatedAt:      now.Add(-12 * time.Hour),
	}

	s.tasks[rootTask.ID] = &rootTask
	s.tasks[childTask.ID] = &childTask
	s.tasks[finalTask.ID] = &finalTask
	s.taskOrder = append(s.taskOrder, rootTask.ID, childTask.ID, finalTask.ID)

	dependency := domain.TaskDependency{
		ID:              newID(),
		TaskID:          finalTask.ID,
		DependsOnTaskID: childTask.ID,
		CreatedAt:       now.Add(-10 * time.Hour),
	}
	s.deps[dependency.ID] = &dependency
	s.depsByTask[finalTask.ID] = append(s.depsByTask[finalTask.ID], dependency.ID)

	generation := domain.GenerationDetails{
		Generation: domain.Generation{
			ID:        newID(),
			GoalID:    goal.ID,
			Status:    domain.GenerationStatusSuccess,
			ModelName: "goalmind-synth-v1",
			CreatedAt: now.Add(-15 * time.Hour),
		},
		Prompt:        fmt.Sprintf("Create a practical breakdown for %s", title),
		RawAIResponse: "Synthetic seed generation used for local development.",
	}
	s.generations[generation.ID] = &generation
	s.generationsByGoal[goal.ID] = append(s.generationsByGoal[goal.ID], generation.ID)

	version := domain.PlanVersion{
		ID:            newID(),
		GoalID:        goal.ID,
		GenerationID:  generation.ID,
		VersionNumber: 1,
		IsActive:      true,
		CreatedAt:     now.Add(-15 * time.Hour),
	}
	s.versions[version.ID] = &version
	s.versionsByGoal[goal.ID] = append(s.versionsByGoal[goal.ID], version.ID)

	return goal.ID
}

func (s *MemoryStore) userOwnsGoalLocked(userID, goalID string) bool {
	goal, ok := s.goals[goalID]
	return ok && goal.UserID == userID
}

func (s *MemoryStore) tasksForGoalLocked(goalID string) []domain.Task {
	result := make([]domain.Task, 0)
	for _, taskID := range s.taskOrder {
		task, ok := s.tasks[taskID]
		if !ok || task.GoalID != goalID {
			continue
		}
		result = append(result, *task)
	}

	slices.SortFunc(result, func(a, b domain.Task) int {
		if a.OrderIndex != b.OrderIndex {
			return a.OrderIndex - b.OrderIndex
		}
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	return result
}

func (s *MemoryStore) goalsForUserLocked(userID string) []domain.Goal {
	result := make([]domain.Goal, 0)
	for _, goalID := range s.goalOrder {
		if goal, ok := s.goals[goalID]; ok && goal.UserID == userID {
			result = append(result, goal.Goal)
		}
	}
	return result
}

func (s *MemoryStore) touchGoalLocked(goalID string) {
	if goal, ok := s.goals[goalID]; ok {
		goal.UpdatedAt = time.Now().UTC()
		if goal.Status == domain.GoalStatusDraft {
			goal.Status = domain.GoalStatusActive
		}
	}
}

func (s *MemoryStore) createGeneratedTaskLocked(goalID string, parentTaskID *string, generated domain.GeneratedPlanTask, orderIndex *int, now time.Time, taskIDsByTitle map[string]string, pendingDependencies map[string][]string) {
	title := strings.TrimSpace(generated.Title)
	if title == "" {
		title = "Generated task"
	}

	task := domain.Task{
		ID:             newID(),
		GoalID:         goalID,
		ParentTaskID:   parentTaskID,
		Title:          title,
		Description:    strings.TrimSpace(generated.Description),
		Status:         domain.TaskStatusTodo,
		Priority:       normalizePriority(generated.Priority),
		EstimatedHours: generated.EstimatedHours,
		Deadline:       normalizeDatePointer(generated.Deadline),
		OrderIndex:     *orderIndex,
		Source:         domain.TaskSourceAI,
		EditedByUser:   false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if task.EstimatedHours < 0 {
		task.EstimatedHours = 0
	}

	s.tasks[task.ID] = &task
	s.taskOrder = append(s.taskOrder, task.ID)
	taskIDsByTitle[task.Title] = task.ID
	pendingDependencies[task.ID] = append([]string(nil), generated.DependsOn...)
	(*orderIndex)++

	parentID := task.ID
	for _, child := range generated.Children {
		s.createGeneratedTaskLocked(goalID, &parentID, child, orderIndex, now, taskIDsByTitle, pendingDependencies)
	}
}

func (s *MemoryStore) ensureBaselineVersionLocked(goalID string) string {
	if versionIDs := s.versionsByGoal[goalID]; len(versionIDs) > 0 {
		for _, versionID := range versionIDs {
			if version, ok := s.versions[versionID]; ok && version.IsActive {
				return versionID
			}
		}
		return versionIDs[len(versionIDs)-1]
	}

	generation := domain.GenerationDetails{
		Generation: domain.Generation{
			ID:        newID(),
			GoalID:    goalID,
			Status:    domain.GenerationStatusSuccess,
			ModelName: "goalmind-synth-v1",
			CreatedAt: time.Now().UTC(),
		},
		Prompt:        "Initialize baseline plan version.",
		RawAIResponse: "Baseline plan version created automatically.",
	}
	s.generations[generation.ID] = &generation
	s.generationsByGoal[goalID] = append(s.generationsByGoal[goalID], generation.ID)

	version := domain.PlanVersion{
		ID:            newID(),
		GoalID:        goalID,
		GenerationID:  generation.ID,
		VersionNumber: 1,
		IsActive:      true,
		CreatedAt:     time.Now().UTC(),
	}
	s.versions[version.ID] = &version
	s.versionsByGoal[goalID] = append(s.versionsByGoal[goalID], version.ID)
	return version.ID
}

func (s *MemoryStore) findUserGoalTitleConflictLocked(userID, title, excludeGoalID string) bool {
	for _, goalID := range s.goalOrder {
		goal, ok := s.goals[goalID]
		if !ok || goal.UserID != userID || goal.ID == excludeGoalID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(goal.Title), strings.TrimSpace(title)) {
			return true
		}
	}
	return false
}

func (s *MemoryStore) deleteTaskCascadeLocked(taskID string) {
	for _, task := range s.tasks {
		if task.ParentTaskID != nil && *task.ParentTaskID == taskID {
			s.deleteTaskCascadeLocked(task.ID)
		}
	}

	delete(s.tasks, taskID)
	s.taskOrder = removeString(s.taskOrder, taskID)
	delete(s.depsByTask, taskID)
	for depID, dep := range s.deps {
		if dep.TaskID == taskID || dep.DependsOnTaskID == taskID {
			delete(s.deps, depID)
		}
	}
}

func buildTaskTree(tasks []domain.Task) []domain.TaskTreeNode {
	childrenByParent := make(map[string][]domain.Task)
	roots := make([]domain.Task, 0)

	for _, task := range tasks {
		if task.ParentTaskID == nil || *task.ParentTaskID == "" {
			roots = append(roots, task)
			continue
		}
		childrenByParent[*task.ParentTaskID] = append(childrenByParent[*task.ParentTaskID], task)
	}

	var build func(task domain.Task) domain.TaskTreeNode
	build = func(task domain.Task) domain.TaskTreeNode {
		children := childrenByParent[task.ID]
		slices.SortFunc(children, func(a, b domain.Task) int {
			if a.OrderIndex != b.OrderIndex {
				return a.OrderIndex - b.OrderIndex
			}
			return a.CreatedAt.Compare(b.CreatedAt)
		})

		node := domain.TaskTreeNode{Task: task, Children: make([]domain.TaskTreeNode, 0, len(children))}
		for _, child := range children {
			node.Children = append(node.Children, build(child))
		}
		return node
	}

	slices.SortFunc(roots, func(a, b domain.Task) int {
		if a.OrderIndex != b.OrderIndex {
			return a.OrderIndex - b.OrderIndex
		}
		return a.CreatedAt.Compare(b.CreatedAt)
	})

	result := make([]domain.TaskTreeNode, 0, len(roots))
	for _, root := range roots {
		result = append(result, build(root))
	}
	return result
}

func (s *MemoryStore) getQualityLocked(goalID string) (domain.QualityResponse, error) {
	goal := s.goals[goalID]
	if goal == nil {
		return domain.QualityResponse{}, ErrNotFound
	}
	response := domain.QualityResponse{GoalID: goalID}
	score := 35.0

	if context, ok := s.goalContext[goalID]; ok {
		score += 20
		if len(context.Preferences) > 0 {
			score += 5
		}
	}

	tasks := s.tasksForGoalLocked(goalID)
	if len(tasks) >= 3 {
		score += 20
	} else if len(tasks) > 0 {
		score += 10
	}

	hasDeadline := false
	for _, task := range tasks {
		if task.Deadline != nil {
			hasDeadline = true
			break
		}
	}
	if hasDeadline {
		score += 10
	}

	if len(s.depsByTask) > 0 {
		score += 10
	}

	response.Score = math.Min(round2(score), 100)
	return response, nil
}

func (s *MemoryStore) getProgressLocked(goalID string) (domain.ProgressResponse, error) {
	if s.goals[goalID] == nil {
		return domain.ProgressResponse{}, ErrNotFound
	}

	response := domain.ProgressResponse{GoalID: goalID}
	allTasks := s.tasksForGoalLocked(goalID)
	parentIDs := make(map[string]bool)
	for _, task := range allTasks {
		if task.ParentTaskID != nil {
			parentIDs[*task.ParentTaskID] = true
		}
	}
	tasks := make([]domain.Task, 0, len(allTasks))
	for _, task := range allTasks {
		if !parentIDs[task.ID] {
			tasks = append(tasks, task)
		}
	}
	response.TotalTasks = len(tasks)

	for _, task := range tasks {
		response.TotalEstimatedHours += task.EstimatedHours
		if task.Status == domain.TaskStatusDone {
			response.CompletedTasks++
			response.CompletedEstimatedHours += task.EstimatedHours
		}
	}

	if response.TotalTasks > 0 {
		response.SimpleProgressPercent = round2(float64(response.CompletedTasks) / float64(response.TotalTasks) * 100)
	}
	if response.TotalEstimatedHours > 0 {
		response.WeightedProgressPercent = round2(response.CompletedEstimatedHours / response.TotalEstimatedHours * 100)
	}

	return response, nil
}

func (s *MemoryStore) getFeasibilityLocked(goalID string) (domain.FeasibilityResponse, error) {
	goal := s.goals[goalID]
	if goal == nil {
		return domain.FeasibilityResponse{}, ErrNotFound
	}

	response := domain.FeasibilityResponse{GoalID: goalID}
	tasks := s.tasksForGoalLocked(goalID)
	for _, task := range tasks {
		response.RequiredHours += task.EstimatedHours
	}

	if goal.Deadline == nil {
		response.Status = domain.FeasibilityStatusUnknown
		return response, nil
	}

	deadlineTime, err := time.Parse("2006-01-02", *goal.Deadline)
	if err != nil {
		response.Status = domain.FeasibilityStatusUnknown
		return response, nil
	}

	hoursUntilDeadline := deadlineTime.Sub(time.Now().UTC()).Hours()
	if hoursUntilDeadline < 0 {
		hoursUntilDeadline = 0
	}
	response.WeeksUntilDeadline = int(math.Ceil(hoursUntilDeadline / (24 * 7)))
	response.AvailableHours = round2(float64(response.WeeksUntilDeadline) * goal.AvailableHoursPerWeek)

	switch {
	case response.RequiredHours == 0:
		response.Status = domain.FeasibilityStatusUnknown
	case response.AvailableHours >= response.RequiredHours:
		response.Status = domain.FeasibilityStatusRealistic
	case response.AvailableHours*1.2 >= response.RequiredHours:
		response.Status = domain.FeasibilityStatusRisky
	default:
		response.Status = domain.FeasibilityStatusUnrealistic
	}
	return response, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizePriority(priority domain.GoalPriority) domain.GoalPriority {
	switch priority {
	case domain.GoalPriorityLow, domain.GoalPriorityMedium, domain.GoalPriorityHigh:
		return priority
	default:
		return domain.GoalPriorityMedium
	}
}

func normalizeGoalStatus(status domain.GoalStatus) domain.GoalStatus {
	switch status {
	case domain.GoalStatusDraft, domain.GoalStatusActive, domain.GoalStatusCompleted, domain.GoalStatusArchived:
		return status
	default:
		return domain.GoalStatusDraft
	}
}

func normalizeTaskStatus(status domain.TaskStatus) domain.TaskStatus {
	switch status {
	case domain.TaskStatusTodo, domain.TaskStatusDone:
		return status
	default:
		return domain.TaskStatusTodo
	}
}

func normalizeDatePointer(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func comparePassword(hash, password string) bool {
	computed := hashPassword(password)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(computed)) == 1
}

func newToken() string {
	return "gm_" + strings.ReplaceAll(newID(), "-", "")
}

func newID() string {
	var data [16]byte
	_, _ = rand.Read(data[:])

	data[6] = (data[6] & 0x0f) | 0x40
	data[8] = (data[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		data[0:4],
		data[4:6],
		data[6:8],
		data[8:10],
		data[10:16],
	)
}

func nextOrderIndex(tasks []domain.Task) int {
	max := 0
	for _, task := range tasks {
		if task.OrderIndex > max {
			max = task.OrderIndex
		}
	}
	return max + 1
}

func removeString(values []string, target string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != target {
			result = append(result, value)
		}
	}
	return result
}

func mergeStoreQuestions(existing []domain.ClarifyingQuestion, additions ...domain.ClarifyingQuestion) []domain.ClarifyingQuestion {
	seen := make(map[string]bool, len(existing)+len(additions))
	result := make([]domain.ClarifyingQuestion, 0, len(existing)+len(additions))
	for _, question := range existing {
		seen[question.ID] = true
		result = append(result, question)
	}
	for _, question := range additions {
		if !seen[question.ID] {
			result = append(result, question)
		}
	}
	return result
}

func validAnswerValue(question domain.ClarifyingQuestion, raw json.RawMessage) bool {
	switch question.Type {
	case domain.QuestionTypeText, domain.QuestionTypeTextarea, domain.QuestionTypeDate, domain.QuestionTypeSingleSelect:
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return false
		}
		return !question.Required || strings.TrimSpace(value) != ""
	case domain.QuestionTypeNumber:
		var value float64
		return json.Unmarshal(raw, &value) == nil
	case domain.QuestionTypeBoolean:
		var value bool
		return json.Unmarshal(raw, &value) == nil
	case domain.QuestionTypeMultiSelect:
		var value []string
		if err := json.Unmarshal(raw, &value); err != nil {
			return false
		}
		if question.Required && len(value) == 0 {
			return false
		}
		return true
	default:
		return false
	}
}

func maskAPIKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 8 {
		return "****"
	}
	return trimmed[:4] + "..." + trimmed[len(trimmed)-4:]
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ryoshimaru/MindMap/internal/domain"
	"github.com/ryoshimaru/MindMap/internal/security"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) GetUserProfile(userID string) (domain.UserProfile, error) {
	var profile domain.UserProfile
	err := s.db.QueryRow(`
		select user_id, age, occupation, free_hours_per_week, available_budget, constraints, created_at, updated_at
		from user_profiles where user_id = $1
	`, userID).Scan(&profile.UserID, &profile.Age, &profile.Occupation, &profile.FreeHoursPerWeek, &profile.AvailableBudget, &profile.Constraints, &profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.UserProfile{}, ErrNotFound
	}
	return profile, err
}

func (s *PostgresStore) SaveUserProfile(userID string, req domain.UpsertUserProfileRequest) (domain.UserProfile, error) {
	if req.Age < 14 || req.Age > 120 || strings.TrimSpace(req.Occupation) == "" || req.FreeHoursPerWeek <= 0 || req.FreeHoursPerWeek > 168 || req.AvailableBudget < 0 {
		return domain.UserProfile{}, ErrValidation
	}
	_, err := s.db.Exec(`
		insert into user_profiles (user_id, age, occupation, free_hours_per_week, available_budget, constraints)
		values ($1,$2,$3,$4,$5,$6)
		on conflict (user_id) do update set age=excluded.age, occupation=excluded.occupation,
			free_hours_per_week=excluded.free_hours_per_week, available_budget=excluded.available_budget,
			constraints=excluded.constraints, updated_at=now()
	`, userID, req.Age, strings.TrimSpace(req.Occupation), req.FreeHoursPerWeek, req.AvailableBudget, strings.TrimSpace(req.Constraints))
	if err != nil {
		return domain.UserProfile{}, err
	}
	return s.GetUserProfile(userID)
}

func (s *PostgresStore) RegisterUser(req domain.RegisterRequest) (domain.AuthResponse, error) {
	email := normalizeEmail(req.Email)
	if email == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Name) == "" {
		return domain.AuthResponse{}, ErrValidation
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.AuthResponse{}, err
	}
	defer rollback(tx)

	user := domain.User{
		ID:        newID(),
		Email:     email,
		Name:      strings.TrimSpace(req.Name),
		Provider:  domain.AuthProviderLocal,
		CreatedAt: time.Now().UTC(),
	}
	_, err = tx.ExecContext(ctx, `
		insert into users (id, email, name, avatar_url, provider, password_hash, created_at)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID, user.Email, user.Name, user.AvatarURL, user.Provider, hashPassword(req.Password), user.CreatedAt)
	if isUniqueViolation(err) {
		return domain.AuthResponse{}, ErrConflict
	}
	if err != nil {
		return domain.AuthResponse{}, err
	}

	token := newToken()
	if _, err := tx.ExecContext(ctx, `insert into sessions (token, user_id, created_at) values ($1, $2, $3)`, token, user.ID, time.Now().UTC()); err != nil {
		return domain.AuthResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.AuthResponse{}, err
	}
	return domain.AuthResponse{AccessToken: token, User: user}, nil
}

func (s *PostgresStore) Login(req domain.LoginRequest) (domain.AuthResponse, error) {
	email := normalizeEmail(req.Email)
	if email == "" || strings.TrimSpace(req.Password) == "" {
		return domain.AuthResponse{}, ErrValidation
	}

	var user domain.User
	var passwordHash sql.NullString
	err := s.db.QueryRow(`
		select id, email, name, avatar_url, provider, password_hash, created_at
		from users
		where email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Provider, &passwordHash, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AuthResponse{}, ErrInvalidAuth
	}
	if err != nil {
		return domain.AuthResponse{}, err
	}
	if !passwordHash.Valid || !comparePassword(passwordHash.String, req.Password) {
		return domain.AuthResponse{}, ErrInvalidAuth
	}

	token := newToken()
	if _, err := s.db.Exec(`insert into sessions (token, user_id, created_at) values ($1, $2, $3)`, token, user.ID, time.Now().UTC()); err != nil {
		return domain.AuthResponse{}, err
	}
	return domain.AuthResponse{AccessToken: token, User: user}, nil
}

func (s *PostgresStore) UserByToken(token string) (domain.User, error) {
	var user domain.User
	err := s.db.QueryRow(`
		select u.id, u.email, u.name, u.avatar_url, u.provider, u.created_at
		from sessions s
		join users u on u.id = s.user_id
		where s.token = $1
	`, strings.TrimSpace(token)).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Provider, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, ErrUnauthorized
	}
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *PostgresStore) CreateOAuthUserSession(email, name, avatarURL string, provider domain.AuthProvider) (string, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)
	if email == "" {
		return "", ErrValidation
	}
	if name == "" {
		name = email
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer rollback(tx)

	var userID string
	err = tx.QueryRowContext(ctx, `select id from users where email = $1`, email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		userID = newID()
		var avatar *string
		if strings.TrimSpace(avatarURL) != "" {
			trimmed := strings.TrimSpace(avatarURL)
			avatar = &trimmed
		}
		_, err = tx.ExecContext(ctx, `
			insert into users (id, email, name, avatar_url, provider, password_hash, created_at)
			values ($1, $2, $3, $4, $5, null, $6)
		`, userID, email, name, avatar, provider, time.Now().UTC())
	}
	if err != nil {
		return "", err
	}

	token := newToken()
	if _, err := tx.ExecContext(ctx, `insert into sessions (token, user_id, created_at) values ($1, $2, $3)`, token, userID, time.Now().UTC()); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *PostgresStore) GetAISettings(userID string) (domain.AISettingsResponse, error) {
	var provider, encrypted string
	err := s.db.QueryRow(`select provider, encrypted_api_key from user_ai_settings where user_id = $1`, userID).Scan(&provider, &encrypted)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AISettingsResponse{}, nil
	}
	if err != nil {
		return domain.AISettingsResponse{}, err
	}
	apiKey, err := security.DecryptAPIKey(encrypted)
	if err != nil {
		return domain.AISettingsResponse{}, err
	}
	return domain.AISettingsResponse{Provider: provider, MaskedKey: maskAPIKey(apiKey)}, nil
}

func (s *PostgresStore) GetAISettingsSecret(userID string) (domain.AISettingsRequest, error) {
	var settings domain.AISettingsRequest
	var encrypted string
	err := s.db.QueryRow(`select provider, encrypted_api_key from user_ai_settings where user_id = $1`, userID).Scan(&settings.Provider, &encrypted)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.AISettingsRequest{}, nil
	}
	if err != nil {
		return domain.AISettingsRequest{}, err
	}
	settings.APIKey, err = security.DecryptAPIKey(encrypted)
	return settings, err
}

func (s *PostgresStore) SaveAISettings(userID string, req domain.AISettingsRequest) (domain.AISettingsResponse, error) {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	apiKey := strings.TrimSpace(req.APIKey)
	if provider != string(domain.AIProviderGemini) && provider != string(domain.AIProviderDeepSeek) {
		return domain.AISettingsResponse{}, ErrValidation
	}
	if apiKey == "" {
		return domain.AISettingsResponse{}, ErrValidation
	}
	encrypted, err := security.EncryptAPIKey(apiKey)
	if err != nil {
		return domain.AISettingsResponse{}, err
	}
	_, err = s.db.Exec(`
		insert into user_ai_settings (user_id, provider, encrypted_api_key, created_at, updated_at)
		values ($1, $2, $3, now(), now())
		on conflict (user_id) do update set
			provider = excluded.provider,
			encrypted_api_key = excluded.encrypted_api_key,
			updated_at = now()
	`, userID, provider, encrypted)
	if err != nil {
		return domain.AISettingsResponse{}, err
	}
	return domain.AISettingsResponse{Provider: provider, MaskedKey: maskAPIKey(apiKey)}, nil
}

func (s *PostgresStore) SaveAnalyzedRequest(userID, originalText string, provider domain.AIProviderName, analysis domain.AnalyzeFreeformResponse) (domain.AnalyzeFreeformResponse, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(originalText) == "" {
		return domain.AnalyzeFreeformResponse{}, ErrValidation
	}
	requestID := newID()
	now := time.Now().UTC()
	tracker := analysis.ActivityTracker
	tracker.Stage = "questions_ready"

	_, err := s.db.Exec(`
		insert into planning_requests
			(id, user_id, original_text, provider, interpreted_goal, feasibility, clarifying_questions, answers, activity_tracker, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, '[]', $8, $9, $9)
	`, requestID, userID, strings.TrimSpace(originalText), provider, mustJSON(analysis.InterpretedGoal), mustJSON(analysis.Feasibility), mustJSON(analysis.ClarifyingQuestions), mustJSON(tracker), now)
	if err != nil {
		return domain.AnalyzeFreeformResponse{}, err
	}

	analysis.RequestID = requestID
	analysis.OriginalText = strings.TrimSpace(originalText)
	analysis.ActivityTracker = tracker
	return analysis, nil
}

func (s *PostgresStore) SaveRequestAnswers(userID, requestID string, req domain.SubmitRequestAnswers) (domain.SubmitRequestAnswersResponse, error) {
	if len(req.Answers) == 0 {
		return domain.SubmitRequestAnswersResponse{}, ErrValidation
	}
	record, err := s.GetPlanningRequest(userID, requestID)
	if err != nil {
		return domain.SubmitRequestAnswersResponse{}, err
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
		question, ok := questionsByID[answer.QuestionID]
		if strings.TrimSpace(answer.QuestionID) == "" || len(answer.Value) == 0 || !ok || !validAnswerValue(question, answer.Value) {
			return domain.SubmitRequestAnswersResponse{}, ErrValidation
		}
		answersByQuestion[answer.QuestionID] = answer
	}

	answers := make([]domain.RequestAnswer, 0, len(answersByQuestion))
	for _, question := range record.ClarifyingQuestions {
		if answer, ok := answersByQuestion[question.ID]; ok {
			answers = append(answers, answer)
		}
	}
	record.ActivityTracker.Stage = "answers_saved"
	record.ActivityTracker.Items = append(record.ActivityTracker.Items, domain.ActivitySignal{Label: "Answers saved", Value: fmt.Sprintf("%d", len(answers))})

	_, err = s.db.Exec(`
		update planning_requests
		set answers = $1, activity_tracker = $2, updated_at = now()
		where id = $3 and user_id = $4
	`, mustJSON(answers), mustJSON(record.ActivityTracker), requestID, userID)
	if err != nil {
		return domain.SubmitRequestAnswersResponse{}, err
	}
	return domain.SubmitRequestAnswersResponse{RequestID: requestID, AnswersSaved: len(answers), ActivityTracker: record.ActivityTracker}, nil
}

func (s *PostgresStore) GetPlanningRequest(userID, requestID string) (domain.PlanningRequest, error) {
	var record domain.PlanningRequest
	var interpretedGoal, feasibility, questions, answers, tracker []byte
	err := s.db.QueryRow(`
		select id, user_id, original_text, provider, interpreted_goal, feasibility, clarifying_questions, answers, activity_tracker, created_at, updated_at
		from planning_requests
		where id = $1 and user_id = $2
	`, requestID, userID).Scan(&record.ID, &record.UserID, &record.OriginalText, &record.Provider, &interpretedGoal, &feasibility, &questions, &answers, &tracker, &record.CreatedAt, &record.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PlanningRequest{}, ErrNotFound
	}
	if err != nil {
		return domain.PlanningRequest{}, err
	}
	if err := decodeJSON(interpretedGoal, &record.InterpretedGoal); err != nil {
		return domain.PlanningRequest{}, err
	}
	if err := decodeJSON(feasibility, &record.Feasibility); err != nil {
		return domain.PlanningRequest{}, err
	}
	if err := decodeJSON(questions, &record.ClarifyingQuestions); err != nil {
		return domain.PlanningRequest{}, err
	}
	if err := decodeJSON(answers, &record.Answers); err != nil {
		return domain.PlanningRequest{}, err
	}
	if err := decodeJSON(tracker, &record.ActivityTracker); err != nil {
		return domain.PlanningRequest{}, err
	}
	return record, nil
}

func (s *PostgresStore) SaveRequestEvaluation(userID, requestID string, evaluation domain.EvaluateRequestResponse) (domain.EvaluateRequestResponse, error) {
	record, err := s.GetPlanningRequest(userID, requestID)
	if err != nil {
		return domain.EvaluateRequestResponse{}, err
	}
	if evaluation.ClarifyingQuestions != nil {
		record.ClarifyingQuestions = mergeStoreQuestions(record.ClarifyingQuestions, evaluation.ClarifyingQuestions...)
	}
	evaluation.RequestID = requestID
	evaluation.ClarifyingQuestions = record.ClarifyingQuestions
	evaluation.ActivityTracker.Stage = "feasibility_evaluated"

	_, err = s.db.Exec(`
		update planning_requests
		set feasibility = $1, clarifying_questions = $2, activity_tracker = $3, updated_at = now()
		where id = $4 and user_id = $5
	`, mustJSON(evaluation.Feasibility), mustJSON(record.ClarifyingQuestions), mustJSON(evaluation.ActivityTracker), requestID, userID)
	if err != nil {
		return domain.EvaluateRequestResponse{}, err
	}
	return evaluation, nil
}

func (s *PostgresStore) CreateGeneratedPlan(userID, requestID string, plan domain.GeneratedPlan) (domain.GeneratePlanResponse, error) {
	record, err := s.GetPlanningRequest(userID, requestID)
	if err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	if !record.Feasibility.CanGeneratePlan {
		return domain.GeneratePlanResponse{}, ErrInvalidState
	}
	if strings.TrimSpace(plan.Goal.Title) == "" || len(plan.Tasks) == 0 {
		return domain.GeneratePlanResponse{}, ErrValidation
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	defer rollback(tx)

	now := time.Now().UTC()
	goal := domain.Goal{
		ID:                    newID(),
		Title:                 strings.TrimSpace(plan.Goal.Title),
		Description:           strings.TrimSpace(plan.Goal.Description),
		Category:              strings.TrimSpace(plan.Goal.Category),
		Priority:              normalizePriority(plan.Goal.Priority),
		Status:                domain.GoalStatusActive,
		Deadline:              plan.Goal.Deadline,
		AvailableHoursPerWeek: math.Max(0, plan.Goal.AvailableHoursPerWeek),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if goal.Category == "" {
		goal.Category = "General"
	}
	if _, err := tx.ExecContext(ctx, `
		insert into goals (id, user_id, title, description, category, priority, status, deadline, available_hours_per_week, created_at, updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)
	`, goal.ID, userID, goal.Title, goal.Description, goal.Category, goal.Priority, goal.Status, dateArg(goal.Deadline), goal.AvailableHoursPerWeek, now); err != nil {
		return domain.GeneratePlanResponse{}, err
	}

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
	if _, err := tx.ExecContext(ctx, `
		insert into goal_contexts (id, goal_id, current_level, constraints_text, preferences, expected_result, created_at, updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$7)
	`, context.ID, context.GoalID, context.CurrentLevel, context.ConstraintsText, mustJSON(context.Preferences), context.ExpectedResult, now); err != nil {
		return domain.GeneratePlanResponse{}, err
	}

	order := 0
	taskIDsByTitle := make(map[string]string)
	pendingDependencies := make(map[string][]string)
	for _, task := range plan.Tasks {
		if err := s.insertGeneratedTask(ctx, tx, goal.ID, nil, task, &order, now, taskIDsByTitle, pendingDependencies); err != nil {
			return domain.GeneratePlanResponse{}, err
		}
	}
	for taskID, titles := range pendingDependencies {
		for _, title := range titles {
			dependsOnID, ok := taskIDsByTitle[strings.TrimSpace(title)]
			if !ok || dependsOnID == taskID {
				continue
			}
			if _, err := tx.ExecContext(ctx, `insert into task_dependencies (id, task_id, depends_on_task_id, created_at) values ($1,$2,$3,$4) on conflict do nothing`, newID(), taskID, dependsOnID, now); err != nil {
				return domain.GeneratePlanResponse{}, err
			}
		}
	}

	generationID := newID()
	if _, err := tx.ExecContext(ctx, `
		insert into generations (id, goal_id, request_id, status, model_name, error_message, prompt, raw_ai_response, created_at)
		values ($1,$2,$3,$4,$5,null,$6,$7,$8)
	`, generationID, goal.ID, requestID, domain.GenerationStatusSuccess, "ai-provider", record.OriginalText, "Generated from structured request flow.", now); err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	versionID := newID()
	if _, err := tx.ExecContext(ctx, `
		insert into plan_versions (id, goal_id, generation_id, version_number, is_active, created_at)
		values ($1,$2,$3,1,true,$4)
	`, versionID, goal.ID, generationID, now); err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		insert into roadmap_history (id, goal_id, request_id, generation_id, plan_version_id, feasibility, snapshot, created_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8)
	`, newID(), goal.ID, requestID, generationID, versionID, mustJSON(record.Feasibility), mustJSON(plan), now); err != nil {
		return domain.GeneratePlanResponse{}, err
	}

	record.ActivityTracker.Stage = "plan_generated"
	record.ActivityTracker.Items = append(record.ActivityTracker.Items, domain.ActivitySignal{Label: "Generated goal", Value: goal.Title})
	if _, err := tx.ExecContext(ctx, `update planning_requests set activity_tracker = $1, updated_at = now() where id = $2`, mustJSON(record.ActivityTracker), requestID); err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.GeneratePlanResponse{}, err
	}

	tasks, err := s.ListTasksByGoal(userID, goal.ID)
	if err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	progress, err := s.GetProgress(userID, goal.ID)
	if err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	return domain.GeneratePlanResponse{RequestID: requestID, Feasibility: record.Feasibility, Goal: goal, Tasks: tasks, Progress: progress}, nil
}

func (s *PostgresStore) ReplaceGoalPlan(userID, goalID string, plan domain.GeneratedPlan) (int, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return 0, err
	}
	if len(plan.Tasks) == 0 {
		return 0, ErrValidation
	}
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer rollback(tx)
	if _, err := tx.ExecContext(ctx, `delete from tasks where goal_id=$1`, goalID); err != nil {
		return 0, err
	}
	title := strings.TrimSpace(plan.Goal.Title)
	if title == "" {
		title = "Цель"
	}
	if _, err := tx.ExecContext(ctx, `update goals set title=$1, description=$2, deadline=$3, status=$4, updated_at=now() where id=$5 and user_id=$6`, title, strings.TrimSpace(plan.Goal.Description), dateArg(plan.Goal.Deadline), domain.GoalStatusActive, goalID, userID); err != nil {
		return 0, err
	}
	order := 0
	ids := map[string]string{}
	pending := map[string][]string{}
	now := time.Now().UTC()
	for _, generated := range plan.Tasks {
		if err := s.insertGeneratedTask(ctx, tx, goalID, nil, generated, &order, now, ids, pending); err != nil {
			return 0, err
		}
	}
	for taskID, titles := range pending {
		for _, dependencyTitle := range titles {
			if dependsOnID, ok := ids[strings.TrimSpace(dependencyTitle)]; ok && dependsOnID != taskID {
				if _, err := tx.ExecContext(ctx, `insert into task_dependencies (id,task_id,depends_on_task_id,created_at) values ($1,$2,$3,$4) on conflict do nothing`, newID(), taskID, dependsOnID, now); err != nil {
					return 0, err
				}
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return order, nil
}

func (s *PostgresStore) ListGoals(userID string) []domain.Goal {
	rows, err := s.db.Query(`
		select id, title, description, category, priority, status, deadline, available_hours_per_week, created_at, updated_at
		from goals
		where user_id = $1
		order by updated_at desc
	`, userID)
	if err != nil {
		return []domain.Goal{}
	}
	defer rows.Close()
	goals := []domain.Goal{}
	for rows.Next() {
		goal, err := scanGoal(rows)
		if err == nil {
			goals = append(goals, goal)
		}
	}
	return goals
}

func (s *PostgresStore) CreateGoal(userID string, req domain.CreateGoalRequest) (domain.Goal, error) {
	if strings.TrimSpace(req.Title) == "" {
		return domain.Goal{}, ErrValidation
	}
	now := time.Now().UTC()
	goal := domain.Goal{
		ID:                    newID(),
		Title:                 strings.TrimSpace(req.Title),
		Description:           strings.TrimSpace(req.Description),
		Category:              strings.TrimSpace(req.Category),
		Priority:              normalizePriority(req.Priority),
		Status:                domain.GoalStatusActive,
		Deadline:              req.Deadline,
		AvailableHoursPerWeek: math.Max(0, req.AvailableHoursPerWeek),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if goal.Category == "" {
		goal.Category = "General"
	}
	_, err := s.db.Exec(`
		insert into goals (id, user_id, title, description, category, priority, status, deadline, available_hours_per_week, created_at, updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)
	`, goal.ID, userID, goal.Title, goal.Description, goal.Category, goal.Priority, goal.Status, dateArg(goal.Deadline), goal.AvailableHoursPerWeek, now)
	return goal, err
}

func (s *PostgresStore) GetGoal(userID, goalID string) (domain.Goal, error) {
	row := s.db.QueryRow(`
		select id, title, description, category, priority, status, deadline, available_hours_per_week, created_at, updated_at
		from goals
		where id = $1 and user_id = $2
	`, goalID, userID)
	goal, err := scanGoal(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Goal{}, ErrNotFound
	}
	return goal, err
}

func (s *PostgresStore) UpdateGoal(userID, goalID string, req domain.UpdateGoalRequest) (domain.Goal, error) {
	goal, err := s.GetGoal(userID, goalID)
	if err != nil {
		return domain.Goal{}, err
	}
	if req.Title != nil {
		goal.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		goal.Description = strings.TrimSpace(*req.Description)
	}
	if req.Category != nil {
		goal.Category = strings.TrimSpace(*req.Category)
	}
	if req.Priority != nil {
		goal.Priority = normalizePriority(*req.Priority)
	}
	if req.Status != nil {
		goal.Status = normalizeGoalStatus(*req.Status)
	}
	if req.Deadline != nil {
		goal.Deadline = req.Deadline
	}
	if req.AvailableHoursPerWeek != nil {
		goal.AvailableHoursPerWeek = math.Max(0, *req.AvailableHoursPerWeek)
	}
	goal.UpdatedAt = time.Now().UTC()
	_, err = s.db.Exec(`
		update goals set title=$1, description=$2, category=$3, priority=$4, status=$5, deadline=$6, available_hours_per_week=$7, updated_at=$8
		where id=$9 and user_id=$10
	`, goal.Title, goal.Description, goal.Category, goal.Priority, goal.Status, dateArg(goal.Deadline), goal.AvailableHoursPerWeek, goal.UpdatedAt, goalID, userID)
	return goal, err
}

func (s *PostgresStore) DeleteGoal(userID, goalID string) error {
	res, err := s.db.Exec(`delete from goals where id = $1 and user_id = $2`, goalID, userID)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) GetGoalContext(userID, goalID string) (domain.GoalContext, error) {
	var context domain.GoalContext
	var preferences []byte
	err := s.db.QueryRow(`
		select c.id, c.goal_id, c.current_level, c.constraints_text, c.preferences, c.expected_result, c.created_at, c.updated_at
		from goal_contexts c
		join goals g on g.id = c.goal_id
		where c.goal_id = $1 and g.user_id = $2
	`, goalID, userID).Scan(&context.ID, &context.GoalID, &context.CurrentLevel, &context.ConstraintsText, &preferences, &context.ExpectedResult, &context.CreatedAt, &context.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GoalContext{}, ErrNotFound
	}
	if err != nil {
		return domain.GoalContext{}, err
	}
	return context, decodeJSON(preferences, &context.Preferences)
}

func (s *PostgresStore) CreateGoalContext(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return domain.GoalContext{}, err
	}
	now := time.Now().UTC()
	context := domain.GoalContext{ID: newID(), GoalID: goalID, CurrentLevel: req.CurrentLevel, ConstraintsText: req.ConstraintsText, Preferences: req.Preferences, ExpectedResult: req.ExpectedResult, CreatedAt: now, UpdatedAt: now}
	_, err := s.db.Exec(`
		insert into goal_contexts (id, goal_id, current_level, constraints_text, preferences, expected_result, created_at, updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$7)
	`, context.ID, goalID, context.CurrentLevel, context.ConstraintsText, mustJSON(context.Preferences), context.ExpectedResult, now)
	if isUniqueViolation(err) {
		return domain.GoalContext{}, ErrConflict
	}
	return context, err
}

func (s *PostgresStore) UpdateGoalContext(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return domain.GoalContext{}, err
	}
	now := time.Now().UTC()
	res, err := s.db.Exec(`
		update goal_contexts
		set current_level=$1, constraints_text=$2, preferences=$3, expected_result=$4, updated_at=$5
		where goal_id=$6
	`, req.CurrentLevel, req.ConstraintsText, mustJSON(req.Preferences), req.ExpectedResult, now, goalID)
	if err != nil {
		return domain.GoalContext{}, err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return domain.GoalContext{}, ErrNotFound
	}
	return s.GetGoalContext(userID, goalID)
}

func (s *PostgresStore) ListTasksByGoal(userID, goalID string) ([]domain.TaskTreeNode, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`
		select id, goal_id, parent_task_id, title, description, status, priority, estimated_hours, deadline, order_index, source, edited_by_user, created_at, updated_at
		from tasks
		where goal_id = $1
		order by order_index, created_at
	`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []domain.Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return buildTaskTree(tasks), rows.Err()
}

func (s *PostgresStore) CreateTask(userID, goalID string, req domain.CreateTaskRequest) (domain.Task, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return domain.Task{}, err
	}
	now := time.Now().UTC()
	status := domain.TaskStatusTodo
	if req.Status != nil {
		status = normalizeTaskStatus(*req.Status)
	}
	orderIndex := 0
	if req.OrderIndex != nil {
		orderIndex = *req.OrderIndex
	}
	task := domain.Task{ID: newID(), GoalID: goalID, ParentTaskID: req.ParentTaskID, Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Description), Status: status, Priority: normalizePriority(req.Priority), EstimatedHours: math.Max(0, req.EstimatedHours), Deadline: req.Deadline, OrderIndex: orderIndex, Source: domain.TaskSourceManual, CreatedAt: now, UpdatedAt: now}
	if task.Title == "" {
		return domain.Task{}, ErrValidation
	}
	return task, s.insertTask(task)
}

func (s *PostgresStore) GetTask(userID, taskID string) (domain.Task, error) {
	row := s.db.QueryRow(`
		select t.id, t.goal_id, t.parent_task_id, t.title, t.description, t.status, t.priority, t.estimated_hours, t.deadline, t.order_index, t.source, t.edited_by_user, t.created_at, t.updated_at
		from tasks t
		join goals g on g.id = t.goal_id
		where t.id = $1 and g.user_id = $2
	`, taskID, userID)
	task, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, ErrNotFound
	}
	return task, err
}

func (s *PostgresStore) UpdateTask(userID, taskID string, req domain.UpdateTaskRequest) (domain.Task, error) {
	task, err := s.GetTask(userID, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	if req.Title != nil {
		task.Title = strings.TrimSpace(*req.Title)
	}
	if req.ParentTaskID != nil {
		parent, parentErr := s.GetTask(userID, *req.ParentTaskID)
		if parentErr != nil || parent.GoalID != task.GoalID || parent.ParentTaskID != nil || parent.ID == task.ID {
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
		task.EstimatedHours = math.Max(0, *req.EstimatedHours)
	}
	if req.Deadline != nil {
		task.Deadline = req.Deadline
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
	_, err = s.db.Exec(`
		update tasks set parent_task_id=$1, title=$2, description=$3, status=$4, priority=$5, estimated_hours=$6, deadline=$7, order_index=$8, edited_by_user=$9, updated_at=$10
		where id=$11
	`, task.ParentTaskID, task.Title, task.Description, task.Status, task.Priority, task.EstimatedHours, dateArg(task.Deadline), task.OrderIndex, task.EditedByUser, task.UpdatedAt, task.ID)
	return task, err
}

func (s *PostgresStore) DeleteTask(userID, taskID string) error {
	task, err := s.GetTask(userID, taskID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`delete from tasks where id = $1`, task.ID)
	return err
}

func (s *PostgresStore) UpdateTaskStatus(userID, taskID string, status domain.TaskStatus) (domain.Task, error) {
	status = normalizeTaskStatus(status)
	if status == domain.TaskStatusDone {
		var unfinished int
		err := s.db.QueryRow(`
			select count(*) from task_dependencies d
			join tasks dependency on dependency.id = d.depends_on_task_id
			join tasks task on task.id = d.task_id
			join goals g on g.id = task.goal_id
			where d.task_id = $1 and g.user_id = $2 and dependency.status <> $3
		`, taskID, userID, domain.TaskStatusDone).Scan(&unfinished)
		if err != nil {
			return domain.Task{}, err
		}
		if unfinished > 0 {
			return domain.Task{}, ErrInvalidState
		}
	}
	edited := true
	return s.UpdateTask(userID, taskID, domain.UpdateTaskRequest{Status: &status, EditedByUser: &edited})
}

func (s *PostgresStore) ListDependencies(userID, taskID string) ([]domain.TaskDependency, error) {
	if _, err := s.GetTask(userID, taskID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`select id, task_id, depends_on_task_id, created_at from task_dependencies where task_id = $1 order by created_at`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.TaskDependency{}
	for rows.Next() {
		var dep domain.TaskDependency
		if err := rows.Scan(&dep.ID, &dep.TaskID, &dep.DependsOnTaskID, &dep.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, dep)
	}
	return result, rows.Err()
}

func (s *PostgresStore) CreateDependency(userID, taskID string, req domain.CreateTaskDependencyRequest) (domain.TaskDependency, error) {
	task, err := s.GetTask(userID, taskID)
	if err != nil {
		return domain.TaskDependency{}, err
	}
	target, err := s.GetTask(userID, req.DependsOnTaskID)
	if err != nil {
		return domain.TaskDependency{}, err
	}
	if task.ID == target.ID || task.GoalID != target.GoalID {
		return domain.TaskDependency{}, ErrValidation
	}
	dep := domain.TaskDependency{ID: newID(), TaskID: taskID, DependsOnTaskID: target.ID, CreatedAt: time.Now().UTC()}
	_, err = s.db.Exec(`insert into task_dependencies (id, task_id, depends_on_task_id, created_at) values ($1,$2,$3,$4)`, dep.ID, dep.TaskID, dep.DependsOnTaskID, dep.CreatedAt)
	if isUniqueViolation(err) {
		return domain.TaskDependency{}, ErrConflict
	}
	return dep, err
}

func (s *PostgresStore) DeleteDependency(userID, taskID, dependencyID string) error {
	if _, err := s.GetTask(userID, taskID); err != nil {
		return err
	}
	res, err := s.db.Exec(`delete from task_dependencies where id = $1 and task_id = $2`, dependencyID, taskID)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) GetProgress(userID, goalID string) (domain.ProgressResponse, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return domain.ProgressResponse{}, err
	}
	var response domain.ProgressResponse
	response.GoalID = goalID
	err := s.db.QueryRow(`
		select count(*),
		       count(*) filter (where status = 'DONE'),
		       coalesce(sum(estimated_hours), 0),
		       coalesce(sum(estimated_hours) filter (where status = 'DONE'), 0)
		from tasks
		where goal_id = $1 and not exists (select 1 from tasks child where child.parent_task_id = tasks.id)
	`, goalID).Scan(&response.TotalTasks, &response.CompletedTasks, &response.TotalEstimatedHours, &response.CompletedEstimatedHours)
	if err != nil {
		return domain.ProgressResponse{}, err
	}
	if response.TotalTasks > 0 {
		response.SimpleProgressPercent = float64(response.CompletedTasks) / float64(response.TotalTasks) * 100
	}
	if response.TotalEstimatedHours > 0 {
		response.WeightedProgressPercent = response.CompletedEstimatedHours / response.TotalEstimatedHours * 100
	}
	return response, nil
}

func (s *PostgresStore) GetFeasibility(userID, goalID string) (domain.FeasibilityResponse, error) {
	goal, err := s.GetGoal(userID, goalID)
	if err != nil {
		return domain.FeasibilityResponse{}, err
	}
	progress, err := s.GetProgress(userID, goalID)
	if err != nil {
		return domain.FeasibilityResponse{}, err
	}
	response := domain.FeasibilityResponse{GoalID: goalID, AvailableHours: goal.AvailableHoursPerWeek, RequiredHours: progress.TotalEstimatedHours, Status: domain.FeasibilityStatusUnknown}
	if goal.Deadline != nil {
		if deadline, err := time.Parse("2006-01-02", *goal.Deadline); err == nil {
			weeks := math.Ceil(time.Until(deadline).Hours() / 24 / 7)
			response.WeeksUntilDeadline = int(math.Max(0, weeks))
			available := float64(response.WeeksUntilDeadline) * goal.AvailableHoursPerWeek
			switch {
			case progress.TotalEstimatedHours == 0:
				response.Status = domain.FeasibilityStatusUnknown
				response.Message = "Add estimated hours to tasks to evaluate feasibility."
			case available == 0:
				response.Status = domain.FeasibilityStatusUnrealistic
				response.Message = "No available time remains before the deadline."
			case progress.TotalEstimatedHours <= available:
				response.Status = domain.FeasibilityStatusRealistic
				response.Message = "The plan fits the available time."
			case progress.TotalEstimatedHours <= available*1.2:
				response.Status = domain.FeasibilityStatusRisky
				response.Message = "The plan is tight and may require scope control."
			default:
				response.Status = domain.FeasibilityStatusUnrealistic
				response.Message = "The plan needs more time or less scope."
			}
		}
	}
	return response, nil
}

func (s *PostgresStore) GetQuality(userID, goalID string) (domain.QualityResponse, error) {
	tasks, err := s.flatTasksForGoal(userID, goalID)
	if err != nil {
		return domain.QualityResponse{}, err
	}
	score := 100.0
	details := []string{}
	if len(tasks) == 0 {
		score = 0
		details = append(details, "No tasks exist yet.")
	}
	for _, task := range tasks {
		if task.EstimatedHours == 0 {
			score -= 5
		}
		if strings.TrimSpace(task.Description) == "" {
			score -= 5
		}
	}
	return domain.QualityResponse{GoalID: goalID, Score: math.Max(0, score), Details: details}, nil
}

func (s *PostgresStore) GetMetrics(userID, goalID string) (domain.MetricsResponse, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return domain.MetricsResponse{}, err
	}
	var response domain.MetricsResponse
	response.GoalID = goalID
	err := s.db.QueryRow(`
		select
			count(*) filter (where source = 'AI'),
			count(*) filter (where edited_by_user),
			count(*) filter (where status = 'DONE')
		from tasks where goal_id = $1
	`, goalID).Scan(&response.TasksCreatedByAI, &response.TasksEditedByUser, &response.TasksCompleted)
	if err != nil {
		return domain.MetricsResponse{}, err
	}
	_ = s.db.QueryRow(`select count(*) from generations where goal_id = $1`, goalID).Scan(&response.RegenerationCount)
	if response.TasksCreatedByAI > 0 {
		response.AcceptedTasksPercent = float64(response.TasksCreatedByAI-response.TasksDeletedByUser) / float64(response.TasksCreatedByAI) * 100
	}
	return response, nil
}

func (s *PostgresStore) ListPlanVersions(userID, goalID string) ([]domain.PlanVersion, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`select id, goal_id, generation_id, version_number, is_active, created_at from plan_versions where goal_id=$1 order by version_number desc`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.PlanVersion{}
	for rows.Next() {
		var version domain.PlanVersion
		if err := rows.Scan(&version.ID, &version.GoalID, &version.GenerationID, &version.VersionNumber, &version.IsActive, &version.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, version)
	}
	return result, rows.Err()
}

func (s *PostgresStore) GetPlanVersion(userID, goalID, versionID string) (domain.PlanVersion, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return domain.PlanVersion{}, err
	}
	var version domain.PlanVersion
	err := s.db.QueryRow(`select id, goal_id, generation_id, version_number, is_active, created_at from plan_versions where id=$1 and goal_id=$2`, versionID, goalID).Scan(&version.ID, &version.GoalID, &version.GenerationID, &version.VersionNumber, &version.IsActive, &version.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PlanVersion{}, ErrNotFound
	}
	return version, err
}

func (s *PostgresStore) ActivatePlanVersion(userID, goalID, versionID string) (domain.PlanVersion, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return domain.PlanVersion{}, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return domain.PlanVersion{}, err
	}
	defer rollback(tx)
	if _, err := tx.Exec(`update plan_versions set is_active=false where goal_id=$1`, goalID); err != nil {
		return domain.PlanVersion{}, err
	}
	res, err := tx.Exec(`update plan_versions set is_active=true where id=$1 and goal_id=$2`, versionID, goalID)
	if err != nil {
		return domain.PlanVersion{}, err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return domain.PlanVersion{}, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return domain.PlanVersion{}, err
	}
	return s.GetPlanVersion(userID, goalID, versionID)
}

func (s *PostgresStore) ListGenerations(userID, goalID string) ([]domain.Generation, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`select id, goal_id, status, model_name, error_message, created_at from generations where goal_id=$1 order by created_at desc`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Generation{}
	for rows.Next() {
		var generation domain.Generation
		if err := rows.Scan(&generation.ID, &generation.GoalID, &generation.Status, &generation.ModelName, &generation.ErrorMessage, &generation.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, generation)
	}
	return result, rows.Err()
}

func (s *PostgresStore) GetGeneration(userID, generationID string) (domain.GenerationDetails, error) {
	var details domain.GenerationDetails
	err := s.db.QueryRow(`
		select gen.id, gen.goal_id, gen.status, gen.model_name, gen.error_message, gen.created_at, gen.prompt, gen.raw_ai_response
		from generations gen
		join goals g on g.id = gen.goal_id
		where gen.id=$1 and g.user_id=$2
	`, generationID, userID).Scan(&details.ID, &details.GoalID, &details.Status, &details.ModelName, &details.ErrorMessage, &details.CreatedAt, &details.Prompt, &details.RawAIResponse)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GenerationDetails{}, ErrNotFound
	}
	return details, err
}

func (s *PostgresStore) CreateFeedback(userID, generationID string, req domain.FeedbackRequest) (domain.Feedback, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return domain.Feedback{}, ErrValidation
	}
	if _, err := s.GetGeneration(userID, generationID); err != nil {
		return domain.Feedback{}, err
	}
	feedback := domain.Feedback{ID: newID(), GenerationID: generationID, Rating: req.Rating, Comment: strings.TrimSpace(req.Comment), CreatedAt: time.Now().UTC()}
	_, err := s.db.Exec(`insert into feedback (id, generation_id, rating, comment, created_at) values ($1,$2,$3,$4,$5)`, feedback.ID, feedback.GenerationID, feedback.Rating, feedback.Comment, feedback.CreatedAt)
	return feedback, err
}

func (s *PostgresStore) DecomposeGoal(userID, goalID string, req domain.DecomposeGoalRequest, mode string) (domain.DecomposeGoalResponse, error) {
	return domain.DecomposeGoalResponse{}, ErrValidation
}

func (s *PostgresStore) DecomposeTask(userID, taskID string, req domain.DecomposeGoalRequest) (domain.DecomposeGoalResponse, error) {
	return domain.DecomposeGoalResponse{}, ErrValidation
}

func (s *PostgresStore) insertGeneratedTask(ctx context.Context, tx *sql.Tx, goalID string, parentTaskID *string, generated domain.GeneratedPlanTask, order *int, now time.Time, taskIDsByTitle map[string]string, pendingDependencies map[string][]string) error {
	taskID := newID()
	task := domain.Task{
		ID:             taskID,
		GoalID:         goalID,
		ParentTaskID:   parentTaskID,
		Title:          strings.TrimSpace(generated.Title),
		Description:    strings.TrimSpace(generated.Description),
		Status:         domain.TaskStatusTodo,
		Priority:       normalizePriority(generated.Priority),
		EstimatedHours: math.Max(0, generated.EstimatedHours),
		Deadline:       generated.Deadline,
		OrderIndex:     *order,
		Source:         domain.TaskSourceAI,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if task.Title == "" {
		return ErrValidation
	}
	taskIDsByTitle[task.Title] = task.ID
	pendingDependencies[task.ID] = append([]string(nil), generated.DependsOn...)
	*order = *order + 1
	if _, err := tx.ExecContext(ctx, `
		insert into tasks (id, goal_id, parent_task_id, title, description, status, priority, estimated_hours, deadline, order_index, source, edited_by_user, created_at, updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,false,$12,$12)
	`, task.ID, task.GoalID, task.ParentTaskID, task.Title, task.Description, task.Status, task.Priority, task.EstimatedHours, dateArg(task.Deadline), task.OrderIndex, task.Source, now); err != nil {
		return err
	}
	for _, child := range generated.Children {
		if err := s.insertGeneratedTask(ctx, tx, goalID, &taskID, child, order, now, taskIDsByTitle, pendingDependencies); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) insertTask(task domain.Task) error {
	_, err := s.db.Exec(`
		insert into tasks (id, goal_id, parent_task_id, title, description, status, priority, estimated_hours, deadline, order_index, source, edited_by_user, created_at, updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$13)
	`, task.ID, task.GoalID, task.ParentTaskID, task.Title, task.Description, task.Status, task.Priority, task.EstimatedHours, dateArg(task.Deadline), task.OrderIndex, task.Source, task.EditedByUser, task.CreatedAt)
	return err
}

func (s *PostgresStore) flatTasksForGoal(userID, goalID string) ([]domain.Task, error) {
	if _, err := s.GetGoal(userID, goalID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`
		select id, goal_id, parent_task_id, title, description, status, priority, estimated_hours, deadline, order_index, source, edited_by_user, created_at, updated_at
		from tasks where goal_id=$1
	`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []domain.Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanGoal(row scanner) (domain.Goal, error) {
	var goal domain.Goal
	var deadline sql.NullTime
	err := row.Scan(&goal.ID, &goal.Title, &goal.Description, &goal.Category, &goal.Priority, &goal.Status, &deadline, &goal.AvailableHoursPerWeek, &goal.CreatedAt, &goal.UpdatedAt)
	if deadline.Valid {
		value := deadline.Time.Format("2006-01-02")
		goal.Deadline = &value
	}
	return goal, err
}

func scanTask(row scanner) (domain.Task, error) {
	var task domain.Task
	var parent sql.NullString
	var deadline sql.NullTime
	err := row.Scan(&task.ID, &task.GoalID, &parent, &task.Title, &task.Description, &task.Status, &task.Priority, &task.EstimatedHours, &deadline, &task.OrderIndex, &task.Source, &task.EditedByUser, &task.CreatedAt, &task.UpdatedAt)
	if parent.Valid {
		task.ParentTaskID = &parent.String
	}
	if deadline.Valid {
		value := deadline.Time.Format("2006-01-02")
		task.Deadline = &value
	}
	return task, err
}

func mustJSON(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return payload
}

func decodeJSON(payload []byte, target any) error {
	if len(payload) == 0 {
		return nil
	}
	return json.Unmarshal(payload, target)
}

func rollback(tx *sql.Tx) {
	if tx != nil {
		_ = tx.Rollback()
	}
}

func dateArg(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return strings.TrimSpace(*value)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint")
}

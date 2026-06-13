package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ryoshimaru/MindMap/internal/domain"
	"github.com/ryoshimaru/MindMap/internal/store"
)

type AIService struct {
	store           store.Store
	defaultProvider domain.AIProviderName
	factories       map[domain.AIProviderName]ProviderFactory
}

type Provider interface {
	Name() string
	AnalyzeRequest(text string, now time.Time) (domain.AnalyzeFreeformResponse, error)
	EvaluateRequest(request domain.PlanningRequest) (domain.EvaluateRequestResponse, error)
	GeneratePlan(request domain.PlanningRequest) (domain.GeneratedPlan, error)
}

type ProviderFactory func(apiKey string) Provider

func NewAIService(store store.Store, defaultProvider domain.AIProviderName, factories map[domain.AIProviderName]ProviderFactory) *AIService {
	return &AIService{store: store, defaultProvider: defaultProvider, factories: factories}
}

func (s *AIService) AnalyzeFreeform(userID string, req domain.AnalyzeFreeformRequest) (domain.AnalyzeFreeformResponse, error) {
	if len([]rune(strings.TrimSpace(req.Text))) < 5 || len([]rune(req.Text)) > 2000 {
		return domain.AnalyzeFreeformResponse{}, store.ErrValidation
	}
	profile, err := s.store.GetUserProfile(userID)
	if err != nil {
		return domain.AnalyzeFreeformResponse{}, store.ErrInvalidState
	}
	providerName, provider, err := s.resolveProvider(userID, req.Provider)
	if err != nil {
		return domain.AnalyzeFreeformResponse{}, err
	}
	contextualText := fmt.Sprintf("ЦЕЛЬ ПОЛЬЗОВАТЕЛЯ:\n%s\n\nКОНТЕКСТ ПРОФИЛЯ (не является частью формулировки цели): возраст %d; род занятий: %s; свободное время: %.1f ч/нед.; бюджет: %.2f; ограничения: %s.", req.Text, profile.Age, profile.Occupation, profile.FreeHoursPerWeek, profile.AvailableBudget, profile.Constraints)
	analysis, err := provider.AnalyzeRequest(contextualText, time.Now().UTC())
	if err != nil {
		return domain.AnalyzeFreeformResponse{}, store.ErrValidation
	}
	if len(analysis.ClarifyingQuestions) > 15 {
		analysis.ClarifyingQuestions = analysis.ClarifyingQuestions[:15]
	}
	return s.store.SaveAnalyzedRequest(userID, req.Text, providerName, analysis)
}

func (s *AIService) SaveAnswers(userID, requestID string, req domain.SubmitRequestAnswers) (domain.SubmitRequestAnswersResponse, error) {
	return s.store.SaveRequestAnswers(userID, requestID, req)
}

func (s *AIService) GeneratePlan(userID, requestID string) (domain.GeneratePlanResponse, error) {
	request, err := s.store.GetPlanningRequest(userID, requestID)
	if err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	if !request.Feasibility.CanGeneratePlan {
		return domain.GeneratePlanResponse{}, store.ErrInvalidState
	}

	_, provider, err := s.resolveProvider(userID, providerStringPointer(string(request.Provider)))
	if err != nil {
		return domain.GeneratePlanResponse{}, err
	}
	plan, err := provider.GeneratePlan(request)
	if err != nil {
		return domain.GeneratePlanResponse{}, err
	}

	plan = normalizeGeneratedPlan(plan)
	return s.store.CreateGeneratedPlan(userID, requestID, plan)
}

func (s *AIService) EvaluateRequest(userID, requestID string) (domain.EvaluateRequestResponse, error) {
	request, err := s.store.GetPlanningRequest(userID, requestID)
	if err != nil {
		return domain.EvaluateRequestResponse{}, err
	}
	_, provider, err := s.resolveProvider(userID, providerStringPointer(string(request.Provider)))
	if err != nil {
		return domain.EvaluateRequestResponse{}, err
	}
	evaluation, err := provider.EvaluateRequest(request)
	if err != nil {
		log.Printf("AI request evaluation failed for provider %s, request %s: %v", request.Provider, requestID, err)
		return domain.EvaluateRequestResponse{}, err
	}
	saved, err := s.store.SaveRequestEvaluation(userID, requestID, evaluation)
	if err != nil {
		log.Printf("Saving AI request evaluation failed for request %s: %v", requestID, err)
	}
	return saved, err
}

func (s *AIService) GetSettings(userID string) (domain.AISettingsResponse, error) {
	settings, err := s.store.GetAISettings(userID)
	if err != nil {
		return domain.AISettingsResponse{}, err
	}
	if strings.TrimSpace(settings.Provider) == "" {
		settings.Provider = string(s.defaultProvider)
	}
	return settings, nil
}

func (s *AIService) SaveSettings(userID string, req domain.AISettingsRequest) (domain.AISettingsResponse, error) {
	providerName := domain.AIProviderName(strings.ToLower(strings.TrimSpace(req.Provider)))
	factory, ok := s.factories[providerName]
	if !ok || strings.TrimSpace(req.APIKey) == "" {
		return domain.AISettingsResponse{}, store.ErrValidation
	}
	if _, err := factory(req.APIKey).AnalyzeRequest("Проверка подключения. Верни краткий корректный JSON-анализ цели: изучить основы Go за месяц.", time.Now().UTC()); err != nil {
		log.Printf("AI key validation failed for provider %s: %v", providerName, err)
		return domain.AISettingsResponse{}, store.ErrInvalidAIKey
	}
	return s.store.SaveAISettings(userID, req)
}

func (s *AIService) DecomposeGoal(userID, goalID string, req domain.DecomposeGoalRequest, mode string) (domain.DecomposeGoalResponse, error) {
	if mode != "regenerate" {
		return s.store.DecomposeGoal(userID, goalID, req, mode)
	}
	goal, err := s.store.GetGoal(userID, goalID)
	if err != nil {
		return domain.DecomposeGoalResponse{}, err
	}
	profile, err := s.store.GetUserProfile(userID)
	if err != nil {
		return domain.DecomposeGoalResponse{}, err
	}
	_, provider, err := s.resolveProvider(userID, nil)
	if err != nil {
		return domain.DecomposeGoalResponse{}, err
	}
	planning := domain.PlanningRequest{
		OriginalText:    fmt.Sprintf("Перестрой план цели %q. Описание: %s. Профиль: возраст %d, занятие %s, %.1f свободных часов в неделю, бюджет %.2f, ограничения: %s", goal.Title, goal.Description, profile.Age, profile.Occupation, profile.FreeHoursPerWeek, profile.AvailableBudget, profile.Constraints),
		InterpretedGoal: domain.InterpretedGoal{Title: goal.Title, Description: goal.Description, Category: goal.Category, Deadline: goal.Deadline},
		Feasibility:     domain.RequestFeasibility{Status: domain.RequestFeasibilityFeasible, CanGeneratePlan: true},
	}
	plan, err := provider.GeneratePlan(planning)
	if err != nil {
		return domain.DecomposeGoalResponse{}, err
	}
	plan = normalizeGeneratedPlan(plan)
	count, err := s.store.ReplaceGoalPlan(userID, goalID, plan)
	if err != nil {
		return domain.DecomposeGoalResponse{}, err
	}
	return domain.DecomposeGoalResponse{GoalID: goalID, TasksCreated: count, FeasibilityStatus: domain.FeasibilityStatusRealistic}, nil
}

func (s *AIService) DecomposeTask(userID, taskID string, req domain.DecomposeGoalRequest) (domain.DecomposeGoalResponse, error) {
	return s.store.DecomposeTask(userID, taskID, req)
}

func (s *AIService) ListGenerations(userID, goalID string) ([]domain.Generation, error) {
	return s.store.ListGenerations(userID, goalID)
}

func (s *AIService) GetGeneration(userID, generationID string) (domain.GenerationDetails, error) {
	return s.store.GetGeneration(userID, generationID)
}

func (s *AIService) CreateFeedback(userID, generationID string, req domain.FeedbackRequest) (domain.Feedback, error) {
	return s.store.CreateFeedback(userID, generationID, req)
}

func (s *AIService) resolveProvider(userID string, requested *string) (domain.AIProviderName, Provider, error) {
	providerName := s.defaultProvider
	apiKey := ""

	settings, _ := s.store.GetAISettingsSecret(userID)
	if settings.Provider != "" {
		providerName = domain.AIProviderName(strings.ToLower(strings.TrimSpace(settings.Provider)))
		apiKey = settings.APIKey
	}
	if requested != nil && strings.TrimSpace(*requested) != "" {
		providerName = domain.AIProviderName(strings.ToLower(strings.TrimSpace(*requested)))
	}

	factory, ok := s.factories[providerName]
	if !ok {
		return "", nil, store.ErrValidation
	}
	return providerName, factory(apiKey), nil
}

func providerStringPointer(value string) *string {
	return &value
}

func normalizeGeneratedPlan(plan domain.GeneratedPlan) domain.GeneratedPlan {
	leaves := 0
	stages := make([]domain.GeneratedPlanTask, 0, len(plan.Tasks))
	for _, stage := range plan.Tasks {
		children := make([]domain.GeneratedPlanTask, 0, len(stage.Children))
		for _, child := range stage.Children {
			if leaves >= 100 {
				break
			}
			child.Children = nil
			children = append(children, child)
			leaves++
		}
		if len(children) == 0 && leaves < 100 {
			stage.DependsOn = nil
			children = append(children, domain.GeneratedPlanTask{Title: stage.Title, Description: stage.Description, Priority: stage.Priority, EstimatedHours: stage.EstimatedHours, Deadline: stage.Deadline})
			stage.Description = ""
			leaves++
		}
		stage.Children = children
		stages = append(stages, stage)
	}
	plan.Tasks = stages
	return plan
}

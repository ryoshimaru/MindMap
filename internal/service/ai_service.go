package service

import (
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
	providerName, provider, err := s.resolveProvider(userID, req.Provider)
	if err != nil {
		return domain.AnalyzeFreeformResponse{}, err
	}
	analysis, err := provider.AnalyzeRequest(req.Text, time.Now().UTC())
	if err != nil {
		return domain.AnalyzeFreeformResponse{}, store.ErrValidation
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
		return domain.EvaluateRequestResponse{}, err
	}
	return s.store.SaveRequestEvaluation(userID, requestID, evaluation)
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
	return s.store.SaveAISettings(userID, req)
}

func (s *AIService) DecomposeGoal(userID, goalID string, req domain.DecomposeGoalRequest, mode string) (domain.DecomposeGoalResponse, error) {
	return s.store.DecomposeGoal(userID, goalID, req, mode)
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

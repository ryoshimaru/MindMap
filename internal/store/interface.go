package store

import "github.com/ryoshimaru/MindMap/internal/domain"

type Store interface {
	RegisterUser(req domain.RegisterRequest) (domain.AuthResponse, error)
	Login(req domain.LoginRequest) (domain.AuthResponse, error)
	UserByToken(token string) (domain.User, error)
	CreateOAuthUserSession(email, name, avatarURL string, provider domain.AuthProvider) (string, error)
	GetUserProfile(userID string) (domain.UserProfile, error)
	SaveUserProfile(userID string, req domain.UpsertUserProfileRequest) (domain.UserProfile, error)

	GetAISettings(userID string) (domain.AISettingsResponse, error)
	GetAISettingsSecret(userID string) (domain.AISettingsRequest, error)
	SaveAISettings(userID string, req domain.AISettingsRequest) (domain.AISettingsResponse, error)

	SaveAnalyzedRequest(userID, originalText string, provider domain.AIProviderName, analysis domain.AnalyzeFreeformResponse) (domain.AnalyzeFreeformResponse, error)
	SaveRequestAnswers(userID, requestID string, req domain.SubmitRequestAnswers) (domain.SubmitRequestAnswersResponse, error)
	GetPlanningRequest(userID, requestID string) (domain.PlanningRequest, error)
	SaveRequestEvaluation(userID, requestID string, evaluation domain.EvaluateRequestResponse) (domain.EvaluateRequestResponse, error)
	CreateGeneratedPlan(userID, requestID string, plan domain.GeneratedPlan) (domain.GeneratePlanResponse, error)
	ReplaceGoalPlan(userID, goalID string, plan domain.GeneratedPlan) (int, error)

	ListGoals(userID string) []domain.Goal
	CreateGoal(userID string, req domain.CreateGoalRequest) (domain.Goal, error)
	GetGoal(userID, goalID string) (domain.Goal, error)
	UpdateGoal(userID, goalID string, req domain.UpdateGoalRequest) (domain.Goal, error)
	DeleteGoal(userID, goalID string) error
	GetGoalContext(userID, goalID string) (domain.GoalContext, error)
	CreateGoalContext(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error)
	UpdateGoalContext(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error)

	ListTasksByGoal(userID, goalID string) ([]domain.TaskTreeNode, error)
	CreateTask(userID, goalID string, req domain.CreateTaskRequest) (domain.Task, error)
	GetTask(userID, taskID string) (domain.Task, error)
	UpdateTask(userID, taskID string, req domain.UpdateTaskRequest) (domain.Task, error)
	DeleteTask(userID, taskID string) error
	UpdateTaskStatus(userID, taskID string, status domain.TaskStatus) (domain.Task, error)
	ListDependencies(userID, taskID string) ([]domain.TaskDependency, error)
	CreateDependency(userID, taskID string, req domain.CreateTaskDependencyRequest) (domain.TaskDependency, error)
	DeleteDependency(userID, taskID, dependencyID string) error

	GetProgress(userID, goalID string) (domain.ProgressResponse, error)
	GetFeasibility(userID, goalID string) (domain.FeasibilityResponse, error)
	GetQuality(userID, goalID string) (domain.QualityResponse, error)
	GetMetrics(userID, goalID string) (domain.MetricsResponse, error)

	ListPlanVersions(userID, goalID string) ([]domain.PlanVersion, error)
	GetPlanVersion(userID, goalID, versionID string) (domain.PlanVersion, error)
	ActivatePlanVersion(userID, goalID, versionID string) (domain.PlanVersion, error)
	ListGenerations(userID, goalID string) ([]domain.Generation, error)
	GetGeneration(userID, generationID string) (domain.GenerationDetails, error)
	CreateFeedback(userID, generationID string, req domain.FeedbackRequest) (domain.Feedback, error)

	DecomposeGoal(userID, goalID string, req domain.DecomposeGoalRequest, mode string) (domain.DecomposeGoalResponse, error)
	DecomposeTask(userID, taskID string, req domain.DecomposeGoalRequest) (domain.DecomposeGoalResponse, error)
}

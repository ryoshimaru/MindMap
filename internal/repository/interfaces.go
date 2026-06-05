package repository

import "github.com/ryoshimaru/MindMap/internal/domain"

type UserRepository interface {
	CreateLocal(req domain.RegisterRequest) (domain.UserRecord, error)
	FindByEmail(email string) (domain.UserRecord, error)
	FindByID(userID string) (domain.UserRecord, error)
}

type SessionRepository interface {
	Create(token, userID string) error
	UserIDByToken(token string) (string, error)
	Delete(token string) error
}

type GoalRepository interface {
	List(userID string) ([]domain.Goal, error)
	Create(userID string, req domain.CreateGoalRequest) (domain.Goal, error)
	Get(userID, goalID string) (domain.Goal, error)
	Update(userID, goalID string, req domain.UpdateGoalRequest) (domain.Goal, error)
	Delete(userID, goalID string) error
}

type GoalContextRepository interface {
	Get(userID, goalID string) (domain.GoalContext, error)
	Create(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error)
	Update(userID, goalID string, req domain.UpsertGoalContextRequest) (domain.GoalContext, error)
}

type TaskRepository interface {
	ListTreeByGoal(userID, goalID string) ([]domain.TaskTreeNode, error)
	Create(userID, goalID string, req domain.CreateTaskRequest) (domain.Task, error)
	Get(userID, taskID string) (domain.Task, error)
	Update(userID, taskID string, req domain.UpdateTaskRequest) (domain.Task, error)
	Delete(userID, taskID string) error
	UpdateStatus(userID, taskID string, status domain.TaskStatus) (domain.Task, error)
}

type DependencyRepository interface {
	List(userID, taskID string) ([]domain.TaskDependency, error)
	Create(userID, taskID string, req domain.CreateTaskDependencyRequest) (domain.TaskDependency, error)
	Delete(userID, taskID, dependencyID string) error
}

type PlanningRequestRepository interface {
	SaveAnalysis(userID, originalText string, analysis domain.AnalyzeFreeformResponse) (domain.AnalyzeFreeformResponse, error)
	SaveAnswers(userID, requestID string, req domain.SubmitRequestAnswers) (domain.SubmitRequestAnswersResponse, error)
	Get(userID, requestID string) (domain.PlanningRequest, error)
	CreateGeneratedPlan(userID, requestID string, plan domain.GeneratedPlan) (domain.GeneratePlanResponse, error)
}

type GenerationRepository interface {
	ListByGoal(userID, goalID string) ([]domain.Generation, error)
	Get(userID, generationID string) (domain.GenerationDetails, error)
	CreateFeedback(userID, generationID string, req domain.FeedbackRequest) (domain.Feedback, error)
}

type VersionRepository interface {
	List(userID, goalID string) ([]domain.PlanVersion, error)
	Get(userID, goalID, versionID string) (domain.PlanVersion, error)
	Activate(userID, goalID, versionID string) (domain.PlanVersion, error)
}

type AnalyticsRepository interface {
	Progress(userID, goalID string) (domain.ProgressResponse, error)
	Feasibility(userID, goalID string) (domain.FeasibilityResponse, error)
	Quality(userID, goalID string) (domain.QualityResponse, error)
	Metrics(userID, goalID string) (domain.MetricsResponse, error)
}

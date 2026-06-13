package domain

import (
	"encoding/json"
	"math"
	"time"
)

type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "LOCAL"
	AuthProviderGoogle AuthProvider = "GOOGLE"
)

type GoalPriority string

const (
	GoalPriorityLow    GoalPriority = "LOW"
	GoalPriorityMedium GoalPriority = "MEDIUM"
	GoalPriorityHigh   GoalPriority = "HIGH"
)

type GoalStatus string

const (
	GoalStatusDraft     GoalStatus = "DRAFT"
	GoalStatusActive    GoalStatus = "ACTIVE"
	GoalStatusCompleted GoalStatus = "COMPLETED"
	GoalStatusArchived  GoalStatus = "ARCHIVED"
)

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "TODO"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusDone       TaskStatus = "DONE"
	TaskStatusCancelled  TaskStatus = "CANCELLED"
)

type TaskSource string

const (
	TaskSourceAI     TaskSource = "AI"
	TaskSourceManual TaskSource = "MANUAL"
)

type FeasibilityStatus string

const (
	FeasibilityStatusRealistic   FeasibilityStatus = "REALISTIC"
	FeasibilityStatusRisky       FeasibilityStatus = "RISKY"
	FeasibilityStatusUnrealistic FeasibilityStatus = "UNREALISTIC"
	FeasibilityStatusUnknown     FeasibilityStatus = "UNKNOWN"
)

type GenerationStatus string

const (
	GenerationStatusPending GenerationStatus = "PENDING"
	GenerationStatusSuccess GenerationStatus = "SUCCESS"
	GenerationStatusFailed  GenerationStatus = "FAILED"
)

type DetailLevel string

const (
	DetailLevelLow    DetailLevel = "LOW"
	DetailLevelMedium DetailLevel = "MEDIUM"
	DetailLevelHigh   DetailLevel = "HIGH"
)

type ClarifyingQuestionType string

const (
	QuestionTypeText         ClarifyingQuestionType = "text"
	QuestionTypeTextarea     ClarifyingQuestionType = "textarea"
	QuestionTypeNumber       ClarifyingQuestionType = "number"
	QuestionTypeDate         ClarifyingQuestionType = "date"
	QuestionTypeSingleSelect ClarifyingQuestionType = "single_select"
	QuestionTypeMultiSelect  ClarifyingQuestionType = "multi_select"
	QuestionTypeBoolean      ClarifyingQuestionType = "boolean"
)

type User struct {
	ID        string       `json:"id"`
	Email     string       `json:"email"`
	Name      string       `json:"name"`
	AvatarURL *string      `json:"avatarUrl"`
	Provider  AuthProvider `json:"provider"`
	CreatedAt time.Time    `json:"createdAt"`
}

type UserProfile struct {
	UserID           string    `json:"userId"`
	Age              int       `json:"age"`
	Occupation       string    `json:"occupation"`
	FreeHoursPerWeek float64   `json:"freeHoursPerWeek"`
	AvailableBudget  float64   `json:"availableBudget"`
	Constraints      string    `json:"constraints"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type UpsertUserProfileRequest struct {
	Age              int     `json:"age"`
	Occupation       string  `json:"occupation"`
	FreeHoursPerWeek float64 `json:"freeHoursPerWeek"`
	AvailableBudget  float64 `json:"availableBudget"`
	Constraints      string  `json:"constraints"`
}

type UserRecord struct {
	User
	PasswordHash string
}

type AuthResponse struct {
	AccessToken string `json:"accessToken"`
	User        User   `json:"user"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AnalyzeFreeformRequest struct {
	Text     string  `json:"text"`
	Provider *string `json:"provider,omitempty"`
}

type AIProviderName string

const (
	AIProviderMock     AIProviderName = "mock"
	AIProviderGemini   AIProviderName = "gemini"
	AIProviderDeepSeek AIProviderName = "deepseek"
)

type RequestFeasibilityStatus string

const (
	RequestFeasibilityUnknown            RequestFeasibilityStatus = "UNKNOWN"
	RequestFeasibilityNeedsClarification RequestFeasibilityStatus = "NEEDS_CLARIFICATION"
	RequestFeasibilityFeasible           RequestFeasibilityStatus = "FEASIBLE"
	RequestFeasibilityRisky              RequestFeasibilityStatus = "RISKY"
	RequestFeasibilityInfeasible         RequestFeasibilityStatus = "INFEASIBLE"
	RequestFeasibilityUnsafe             RequestFeasibilityStatus = "UNSAFE"
)

type RequestFeasibility struct {
	Status                 RequestFeasibilityStatus `json:"status"`
	Reason                 string                   `json:"reason"`
	BlockingFactors        []string                 `json:"blockingFactors"`
	Assumptions            []string                 `json:"assumptions"`
	RequiredClarifications []string                 `json:"requiredClarifications"`
	CanGeneratePlan        bool                     `json:"canGeneratePlan"`
	SuggestedGoal          *InterpretedGoal         `json:"suggestedGoal,omitempty"`
}

type InterpretedGoal struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Deadline    *string  `json:"deadline"`
	Constraints []string `json:"constraints"`
}

type ClarifyingQuestionOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type ClarifyingQuestion struct {
	ID          string                     `json:"id"`
	Text        string                     `json:"text"`
	Type        ClarifyingQuestionType     `json:"type"`
	Required    bool                       `json:"required"`
	Options     []ClarifyingQuestionOption `json:"options,omitempty"`
	Min         *float64                   `json:"min,omitempty"`
	Max         *float64                   `json:"max,omitempty"`
	Unit        *string                    `json:"unit,omitempty"`
	Placeholder *string                    `json:"placeholder,omitempty"`
}

type ActivityTracker struct {
	Stage      string           `json:"stage"`
	Confidence int              `json:"confidence"`
	Items      []ActivitySignal `json:"items"`
}

func (tracker *ActivityTracker) UnmarshalJSON(data []byte) error {
	type activityTrackerJSON struct {
		Stage      string           `json:"stage"`
		Confidence float64          `json:"confidence"`
		Items      []ActivitySignal `json:"items"`
	}

	var decoded activityTrackerJSON
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	confidence := decoded.Confidence
	if confidence >= 0 && confidence <= 1 {
		confidence *= 100
	}
	confidence = math.Max(0, math.Min(100, confidence))

	tracker.Stage = decoded.Stage
	tracker.Confidence = int(math.Round(confidence))
	tracker.Items = decoded.Items
	return nil
}

type ActivitySignal struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type AnalyzeFreeformResponse struct {
	RequestID           string               `json:"requestId"`
	OriginalText        string               `json:"originalText"`
	InterpretedGoal     InterpretedGoal      `json:"interpretedGoal"`
	Feasibility         RequestFeasibility   `json:"feasibility"`
	ClarifyingQuestions []ClarifyingQuestion `json:"clarifyingQuestions"`
	ActivityTracker     ActivityTracker      `json:"activityTracker"`
}

type RequestAnswer struct {
	QuestionID string          `json:"questionId"`
	Value      json.RawMessage `json:"value"`
}

type SubmitRequestAnswers struct {
	Answers []RequestAnswer `json:"answers"`
}

type SubmitRequestAnswersResponse struct {
	RequestID       string          `json:"requestId"`
	AnswersSaved    int             `json:"answersSaved"`
	ActivityTracker ActivityTracker `json:"activityTracker"`
}

type EvaluateRequestResponse struct {
	RequestID           string               `json:"requestId"`
	Feasibility         RequestFeasibility   `json:"feasibility"`
	ClarifyingQuestions []ClarifyingQuestion `json:"clarifyingQuestions"`
	ActivityTracker     ActivityTracker      `json:"activityTracker"`
}

type PlanningRequest struct {
	ID                  string               `json:"id"`
	UserID              string               `json:"-"`
	OriginalText        string               `json:"originalText"`
	Provider            AIProviderName       `json:"provider"`
	InterpretedGoal     InterpretedGoal      `json:"interpretedGoal"`
	Feasibility         RequestFeasibility   `json:"feasibility"`
	ClarifyingQuestions []ClarifyingQuestion `json:"clarifyingQuestions"`
	Answers             []RequestAnswer      `json:"answers"`
	ActivityTracker     ActivityTracker      `json:"activityTracker"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
}

type GeneratedPlanTask struct {
	Title          string              `json:"title"`
	Description    string              `json:"description"`
	Priority       GoalPriority        `json:"priority"`
	EstimatedHours float64             `json:"estimatedHours"`
	Deadline       *string             `json:"deadline"`
	DependsOn      []string            `json:"dependsOn,omitempty"`
	Children       []GeneratedPlanTask `json:"children"`
}

type GeneratedPlan struct {
	Goal    CreateGoalRequest        `json:"goal"`
	Context UpsertGoalContextRequest `json:"context"`
	Tasks   []GeneratedPlanTask      `json:"tasks"`
}

type GeneratePlanResponse struct {
	RequestID   string             `json:"requestId"`
	Feasibility RequestFeasibility `json:"feasibility"`
	Goal        Goal               `json:"goal"`
	Tasks       []TaskTreeNode     `json:"tasks"`
	Progress    ProgressResponse   `json:"progress"`
}

type AISettingsRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey"`
}

type AISettingsResponse struct {
	Provider  string `json:"provider"`
	MaskedKey string `json:"maskedKey"`
}

type Goal struct {
	ID                    string       `json:"id"`
	Title                 string       `json:"title"`
	Description           string       `json:"description"`
	Category              string       `json:"category"`
	Priority              GoalPriority `json:"priority"`
	Status                GoalStatus   `json:"status"`
	Deadline              *string      `json:"deadline"`
	AvailableHoursPerWeek float64      `json:"availableHoursPerWeek"`
	CreatedAt             time.Time    `json:"createdAt"`
	UpdatedAt             time.Time    `json:"updatedAt"`
}

type GoalRecord struct {
	Goal
	UserID string
}

type CreateGoalRequest struct {
	Title                 string       `json:"title"`
	Description           string       `json:"description"`
	Category              string       `json:"category"`
	Priority              GoalPriority `json:"priority"`
	Deadline              *string      `json:"deadline"`
	AvailableHoursPerWeek float64      `json:"availableHoursPerWeek"`
}

type UpdateGoalRequest struct {
	Title                 *string       `json:"title"`
	Description           *string       `json:"description"`
	Category              *string       `json:"category"`
	Priority              *GoalPriority `json:"priority"`
	Status                *GoalStatus   `json:"status"`
	Deadline              *string       `json:"deadline"`
	AvailableHoursPerWeek *float64      `json:"availableHoursPerWeek"`
}

type GoalContext struct {
	ID              string    `json:"id"`
	GoalID          string    `json:"goalId"`
	CurrentLevel    string    `json:"currentLevel"`
	ConstraintsText string    `json:"constraintsText"`
	Preferences     []string  `json:"preferences"`
	ExpectedResult  string    `json:"expectedResult"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type UpsertGoalContextRequest struct {
	CurrentLevel    string   `json:"currentLevel"`
	ConstraintsText string   `json:"constraintsText"`
	Preferences     []string `json:"preferences"`
	ExpectedResult  string   `json:"expectedResult"`
}

type Task struct {
	ID             string       `json:"id"`
	GoalID         string       `json:"goalId"`
	ParentTaskID   *string      `json:"parentTaskId"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Status         TaskStatus   `json:"status"`
	Priority       GoalPriority `json:"priority"`
	EstimatedHours float64      `json:"estimatedHours"`
	Deadline       *string      `json:"deadline"`
	OrderIndex     int          `json:"orderIndex"`
	Source         TaskSource   `json:"source"`
	EditedByUser   bool         `json:"editedByUser"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type TaskTreeNode struct {
	Task
	Children []TaskTreeNode `json:"children"`
}

type CreateTaskRequest struct {
	ParentTaskID   *string      `json:"parentTaskId"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Status         *TaskStatus  `json:"status"`
	Priority       GoalPriority `json:"priority"`
	EstimatedHours float64      `json:"estimatedHours"`
	Deadline       *string      `json:"deadline"`
	OrderIndex     *int         `json:"orderIndex"`
}

type UpdateTaskRequest struct {
	ParentTaskID   *string       `json:"parentTaskId"`
	Title          *string       `json:"title"`
	Description    *string       `json:"description"`
	Status         *TaskStatus   `json:"status"`
	Priority       *GoalPriority `json:"priority"`
	EstimatedHours *float64      `json:"estimatedHours"`
	Deadline       *string       `json:"deadline"`
	OrderIndex     *int          `json:"orderIndex"`
	EditedByUser   *bool         `json:"editedByUser"`
}

type UpdateTaskStatusRequest struct {
	Status TaskStatus `json:"status"`
}

type TaskDependency struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"taskId"`
	DependsOnTaskID string    `json:"dependsOnTaskId"`
	CreatedAt       time.Time `json:"createdAt"`
}

type CreateTaskDependencyRequest struct {
	DependsOnTaskID string `json:"dependsOnTaskId"`
}

type DecomposeGoalRequest struct {
	DetailLevel         DetailLevel `json:"detailLevel"`
	IncludeDependencies bool        `json:"includeDependencies"`
	IncludeDeadlines    bool        `json:"includeDeadlines"`
}

type DecomposeGoalResponse struct {
	GoalID            string            `json:"goalId"`
	GenerationID      string            `json:"generationId"`
	PlanVersionID     string            `json:"planVersionId"`
	TasksCreated      int               `json:"tasksCreated"`
	QualityScore      float64           `json:"qualityScore"`
	FeasibilityStatus FeasibilityStatus `json:"feasibilityStatus"`
}

type ProgressResponse struct {
	GoalID                  string  `json:"goalId"`
	TotalTasks              int     `json:"totalTasks"`
	CompletedTasks          int     `json:"completedTasks"`
	TotalEstimatedHours     float64 `json:"totalEstimatedHours"`
	CompletedEstimatedHours float64 `json:"completedEstimatedHours"`
	SimpleProgressPercent   float64 `json:"simpleProgressPercent"`
	WeightedProgressPercent float64 `json:"weightedProgressPercent"`
}

type FeasibilityResponse struct {
	GoalID             string            `json:"goalId"`
	WeeksUntilDeadline int               `json:"weeksUntilDeadline"`
	AvailableHours     float64           `json:"availableHours"`
	RequiredHours      float64           `json:"requiredHours"`
	Status             FeasibilityStatus `json:"status"`
	Message            string            `json:"message"`
}

type QualityResponse struct {
	GoalID  string   `json:"goalId"`
	Score   float64  `json:"score"`
	Details []string `json:"details"`
}

type MetricsResponse struct {
	GoalID               string  `json:"goalId"`
	TasksCreatedByAI     int     `json:"tasksCreatedByAI"`
	TasksEditedByUser    int     `json:"tasksEditedByUser"`
	TasksDeletedByUser   int     `json:"tasksDeletedByUser"`
	TasksCompleted       int     `json:"tasksCompleted"`
	RegenerationCount    int     `json:"regenerationCount"`
	AcceptedTasksPercent float64 `json:"acceptedTasksPercent"`
}

type PlanVersion struct {
	ID            string    `json:"id"`
	GoalID        string    `json:"goalId"`
	GenerationID  string    `json:"generationId"`
	VersionNumber int       `json:"versionNumber"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Generation struct {
	ID           string           `json:"id"`
	GoalID       string           `json:"goalId"`
	Status       GenerationStatus `json:"status"`
	ModelName    string           `json:"modelName"`
	ErrorMessage *string          `json:"errorMessage"`
	CreatedAt    time.Time        `json:"createdAt"`
}

type GenerationDetails struct {
	Generation
	Prompt        string `json:"prompt"`
	RawAIResponse string `json:"rawAiResponse"`
}

type FeedbackRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type Feedback struct {
	ID           string    `json:"id"`
	GenerationID string    `json:"generationId"`
	Rating       int       `json:"rating"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"createdAt"`
}

export type AuthProviderType = "LOCAL" | "GOOGLE";
export type GoalPriority = "LOW" | "MEDIUM" | "HIGH";
export type GoalStatus = "DRAFT" | "ACTIVE" | "COMPLETED" | "ARCHIVED";
export type TaskStatus = "TODO" | "IN_PROGRESS" | "DONE" | "CANCELLED";
export type TaskSource = "AI" | "MANUAL";
export type DetailLevel = "LOW" | "MEDIUM" | "HIGH";
export type FeasibilityStatus =
  | "REALISTIC"
  | "RISKY"
  | "UNREALISTIC"
  | "UNKNOWN";
export type RequestFeasibilityStatus =
  | "UNKNOWN"
  | "NEEDS_CLARIFICATION"
  | "FEASIBLE"
  | "RISKY"
  | "INFEASIBLE"
  | "UNSAFE";
export type GenerationStatus = "PENDING" | "SUCCESS" | "FAILED";
export type ClarifyingQuestionType =
  | "text"
  | "textarea"
  | "number"
  | "date"
  | "single_select"
  | "multi_select"
  | "boolean";
export type ActivityStage =
  | "questions_ready"
  | "answers_saved"
  | "feasibility_evaluated"
  | "plan_generated"
  | "analysis_ready";

export interface ApiResponse<T> {
  data: T;
  meta?: Record<string, unknown>;
}

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown> | null;
  };
}

export interface User {
  id: string;
  email: string;
  name: string;
  avatarUrl: string | null;
  provider: AuthProviderType;
  createdAt: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  accessToken: string;
  user: User;
}

export interface Goal {
  id: string;
  title: string;
  description: string;
  category: string;
  priority: GoalPriority;
  status: GoalStatus;
  deadline: string | null;
  availableHoursPerWeek: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateGoalRequest {
  title: string;
  description: string;
  category: string;
  priority: GoalPriority;
  deadline: string | null;
  availableHoursPerWeek: number;
}

export interface AnalyzeGoalRequest {
  text: string;
  provider?: string;
}

export interface InterpretedGoal {
  title: string;
  description: string;
  category: string;
  deadline: string | null;
  constraints: string[];
}

export interface ClarifyingQuestionOption {
  value: string;
  label: string;
}

export interface ClarifyingQuestion {
  id: string;
  text: string;
  type: ClarifyingQuestionType;
  required: boolean;
  options?: ClarifyingQuestionOption[];
  min?: number | null;
  max?: number | null;
  unit?: string | null;
  placeholder?: string | null;
}

export interface RequestFeasibility {
  status: RequestFeasibilityStatus;
  reason: string;
  blockingFactors: string[];
  assumptions: string[];
  requiredClarifications: string[];
  canGeneratePlan: boolean;
}

export interface ActivitySignal {
  label: string;
  value: string;
}

export interface ActivityTracker {
  stage: ActivityStage;
  confidence: number;
  items: ActivitySignal[];
}

export interface AnalyzeGoalResponse {
  requestId: string;
  originalText: string;
  interpretedGoal: InterpretedGoal;
  feasibility: RequestFeasibility;
  clarifyingQuestions: ClarifyingQuestion[];
  activityTracker: ActivityTracker;
}

export type RequestAnswerValue = string | number | boolean | string[];

export interface RequestAnswer {
  questionId: string;
  value: RequestAnswerValue;
}

export interface SubmitGoalRequestAnswersRequest {
  answers: RequestAnswer[];
}

export interface SubmitGoalRequestAnswersResponse {
  requestId: string;
  answersSaved: number;
  activityTracker: ActivityTracker;
}

export interface EvaluateGoalRequestResponse {
  requestId: string;
  feasibility: RequestFeasibility;
  clarifyingQuestions: ClarifyingQuestion[];
  activityTracker: ActivityTracker;
}

export interface GeneratePlanFromRequestResponse {
  requestId: string;
  feasibility: RequestFeasibility;
  goal: Goal;
  tasks: TaskTreeNode[];
  progress: ProgressResponse;
}

export interface UpdateGoalRequest {
  title?: string;
  description?: string;
  category?: string;
  priority?: GoalPriority;
  status?: GoalStatus;
  deadline?: string | null;
  availableHoursPerWeek?: number;
}

export interface GoalContext {
  id: string;
  goalId: string;
  currentLevel: string;
  constraintsText: string;
  preferences: string[];
  expectedResult: string;
  createdAt: string;
  updatedAt: string;
}

export interface GoalContextInput {
  currentLevel: string;
  constraintsText: string;
  preferences: string[];
  expectedResult: string;
}

export interface Task {
  id: string;
  goalId: string;
  parentTaskId: string | null;
  title: string;
  description: string;
  status: TaskStatus;
  priority: GoalPriority;
  estimatedHours: number;
  deadline: string | null;
  orderIndex: number;
  source: TaskSource;
  editedByUser: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface TaskTreeNode extends Task {
  children: TaskTreeNode[];
}

export interface CreateTaskRequest {
  parentTaskId?: string | null;
  title: string;
  description: string;
  status?: TaskStatus;
  priority: GoalPriority;
  estimatedHours: number;
  deadline?: string | null;
  orderIndex?: number;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: TaskStatus;
  priority?: GoalPriority;
  estimatedHours?: number;
  deadline?: string | null;
  orderIndex?: number;
  editedByUser?: boolean;
}

export interface UpdateTaskStatusRequest {
  status: TaskStatus;
}

export interface TaskDependency {
  id: string;
  taskId: string;
  dependsOnTaskId: string;
  createdAt: string;
}

export interface CreateTaskDependencyRequest {
  dependsOnTaskId: string;
}

export interface DecomposeGoalRequest {
  detailLevel: DetailLevel;
  includeDependencies: boolean;
  includeDeadlines: boolean;
}

export interface DecomposeGoalResponse {
  goalId: string;
  generationId: string;
  planVersionId: string;
  tasksCreated: number;
  qualityScore: number;
  feasibilityStatus: FeasibilityStatus;
}

export interface ProgressResponse {
  goalId: string;
  totalTasks: number;
  completedTasks: number;
  totalEstimatedHours: number;
  completedEstimatedHours: number;
  simpleProgressPercent: number;
  weightedProgressPercent: number;
}

export interface FeasibilityResponse {
  goalId: string;
  weeksUntilDeadline: number;
  availableHours: number;
  requiredHours: number;
  status: FeasibilityStatus;
  message: string;
}

export interface QualityResponse {
  goalId: string;
  score: number;
  details: string[];
}

export interface MetricsResponse {
  goalId: string;
  tasksCreatedByAI: number;
  tasksEditedByUser: number;
  tasksDeletedByUser: number;
  tasksCompleted: number;
  regenerationCount: number;
  acceptedTasksPercent: number;
}

export interface PlanVersion {
  id: string;
  goalId: string;
  generationId: string;
  versionNumber: number;
  isActive: boolean;
  createdAt: string;
}

export interface Generation {
  id: string;
  goalId: string;
  status: GenerationStatus;
  modelName: string;
  errorMessage: string | null;
  createdAt: string;
}

export interface GenerationDetails extends Generation {
  prompt: string;
  rawAiResponse: string;
}

export interface FeedbackRequest {
  rating: number;
  comment: string;
}

export interface Feedback {
  id: string;
  generationId: string;
  rating: number;
  comment: string;
  createdAt: string;
}

export type AIProviderName = "gemini" | "deepseek" | "mock";

export interface AISettingsRequest {
  provider: Exclude<AIProviderName, "mock">;
  apiKey: string;
}

export interface AISettingsResponse {
  provider: AIProviderName;
  maskedKey: string;
}

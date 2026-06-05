import { ApiClientError } from "./client";
import type {
  AuthResponse,
  CreateGoalRequest,
  CreateTaskDependencyRequest,
  CreateTaskRequest,
  DecomposeGoalRequest,
  DecomposeGoalResponse,
  Feedback,
  FeedbackRequest,
  FeasibilityResponse,
  Generation,
  GenerationDetails,
  Goal,
  GoalContext,
  GoalContextInput,
  AISettingsRequest,
  AISettingsResponse,
  LoginRequest,
  MetricsResponse,
  PlanVersion,
  ProgressResponse,
  QualityResponse,
  RegisterRequest,
  Task,
  TaskDependency,
  TaskTreeNode,
  UpdateGoalRequest,
  UpdateTaskRequest,
  UpdateTaskStatusRequest,
  User
} from "./types";

interface GoalAuditMetrics {
  tasksDeletedByUser: number;
  regenerationCount: number;
}

const DAY_MS = 24 * 60 * 60 * 1000;

const DEMO_TOKEN = "demo-goalmind-token";
const DEMO_USER_ID = "1cb5df19-0789-4f74-a573-5808479c0566";
const GOAL_ONE_ID = "8cb0c04f-1f22-4827-8e54-6a00cf1c8b79";
const GOAL_TWO_ID = "20a00ebc-657b-4219-a36d-35cfc962d4dd";

function nowIso() {
  return new Date().toISOString();
}

function dateInDays(daysFromNow: number) {
  return new Date(Date.now() + daysFromNow * DAY_MS).toISOString().slice(0, 10);
}

function createUuid() {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }

  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (character) => {
    const randomValue = Math.floor(Math.random() * 16);
    const value = character === "x" ? randomValue : (randomValue & 0x3) | 0x8;
    return value.toString(16);
  });
}

function cloneValue<T>(value: T): T {
  if (value === undefined) {
    return value;
  }

  return JSON.parse(JSON.stringify(value)) as T;
}

function respond<T>(producer: () => T, delay = 220): Promise<T> {
  return new Promise((resolve, reject) => {
    window.setTimeout(() => {
      try {
        resolve(cloneValue(producer()));
      } catch (error) {
        reject(error);
      }
    }, delay);
  });
}

function raise(status: number, code: string, message: string): never {
  throw new ApiClientError(message, status, code);
}

const demoUser: User = {
  id: DEMO_USER_ID,
  email: "demo@goalmind.app",
  name: "Anna Petrova",
  avatarUrl: null,
  provider: "LOCAL",
  createdAt: "2026-04-04T09:00:00.000Z"
};

let currentUser: User = demoUser;
let mockAISettings: AISettingsResponse = {
  provider: "mock",
  maskedKey: ""
};

let goals: Goal[] = [
  {
    id: GOAL_ONE_ID,
    title: "Launch GoalMind diploma MVP",
    description:
      "Prepare a polished demo version of the GoalMind web application for thesis defense.",
    category: "Education Product",
    priority: "HIGH",
    status: "ACTIVE",
    deadline: dateInDays(42),
    availableHoursPerWeek: 18,
    createdAt: "2026-04-12T10:00:00.000Z",
    updatedAt: "2026-05-02T16:40:00.000Z"
  },
  {
    id: GOAL_TWO_ID,
    title: "Prepare thesis defense presentation",
    description:
      "Organize the narrative, slides, and rehearsal plan for the final diploma presentation.",
    category: "Academic Planning",
    priority: "MEDIUM",
    status: "DRAFT",
    deadline: dateInDays(21),
    availableHoursPerWeek: 8,
    createdAt: "2026-04-24T11:20:00.000Z",
    updatedAt: "2026-04-29T13:10:00.000Z"
  }
];

let goalContexts: Record<string, GoalContext> = {
  [GOAL_ONE_ID]: {
    id: "0ddbb6df-a9ea-44f5-81c4-8f4f2c492c57",
    goalId: GOAL_ONE_ID,
    currentLevel: "The base dashboard exists, but AI planning, versioning, and analytics need polish.",
    constraintsText:
      "The backend is still under development, so the frontend must support demo mode and a clean API contract.",
    preferences: ["Responsive UI", "Professional visual tone", "Minimal learning curve"],
    expectedResult:
      "A complete, diploma-ready frontend experience with typed API integration points and clear documentation.",
    createdAt: "2026-04-12T10:00:00.000Z",
    updatedAt: "2026-05-02T16:40:00.000Z"
  },
  [GOAL_TWO_ID]: {
    id: "6c26342f-c25a-4308-ac1b-ef266076cb1a",
    goalId: GOAL_TWO_ID,
    currentLevel: "Presentation outline drafted, but the final structure is still loose.",
    constraintsText: "Slides should stay concise, visually clean, and easy to explain live.",
    preferences: ["Strong narrative arc", "Time-boxed rehearsal", "Clear metrics"],
    expectedResult:
      "A confident defense presentation supported by a clear problem statement and demo flow.",
    createdAt: "2026-04-24T11:20:00.000Z",
    updatedAt: "2026-04-29T13:10:00.000Z"
  }
};

let tasks: Task[] = [
  {
    id: "dca1dc94-4b1a-4da7-a5a1-0bfc0c69dcb9",
    goalId: GOAL_ONE_ID,
    parentTaskId: null,
    title: "Stage 1. Product framing",
    description: "Clarify scope, success criteria, and target demo storyline.",
    status: "IN_PROGRESS",
    priority: "HIGH",
    estimatedHours: 10,
    deadline: dateInDays(7),
    orderIndex: 1,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-13T09:00:00.000Z",
    updatedAt: "2026-04-30T15:00:00.000Z"
  },
  {
    id: "aa1d23cd-2743-4d12-b65f-718ce0dc2a1a",
    goalId: GOAL_ONE_ID,
    parentTaskId: "dca1dc94-4b1a-4da7-a5a1-0bfc0c69dcb9",
    title: "Define MVP boundaries",
    description: "Choose which modules must be ready for thesis demonstration.",
    status: "DONE",
    priority: "HIGH",
    estimatedHours: 4,
    deadline: dateInDays(4),
    orderIndex: 1,
    source: "AI",
    editedByUser: true,
    createdAt: "2026-04-13T09:20:00.000Z",
    updatedAt: "2026-04-28T11:00:00.000Z"
  },
  {
    id: "cd6474a6-6d4c-4f3b-8dc8-d4135875d1e0",
    goalId: GOAL_ONE_ID,
    parentTaskId: "dca1dc94-4b1a-4da7-a5a1-0bfc0c69dcb9",
    title: "Gather user context signals",
    description: "List constraints, preferences, and expected outcomes to support decomposition.",
    status: "IN_PROGRESS",
    priority: "MEDIUM",
    estimatedHours: 6,
    deadline: dateInDays(6),
    orderIndex: 2,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-13T09:30:00.000Z",
    updatedAt: "2026-05-01T10:00:00.000Z"
  },
  {
    id: "fc1118f6-b0dc-43cb-8c9d-46c889474e55",
    goalId: GOAL_ONE_ID,
    parentTaskId: null,
    title: "Stage 2. Frontend implementation",
    description: "Build the core pages, plan visualization, and client-side state handling.",
    status: "IN_PROGRESS",
    priority: "HIGH",
    estimatedHours: 26,
    deadline: dateInDays(20),
    orderIndex: 2,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-14T10:00:00.000Z",
    updatedAt: "2026-05-02T15:20:00.000Z"
  },
  {
    id: "a7516263-5512-4ab3-b377-3644f8640f97",
    goalId: GOAL_ONE_ID,
    parentTaskId: "fc1118f6-b0dc-43cb-8c9d-46c889474e55",
    title: "Design dashboard and forms",
    description: "Create the layout, navigation, auth pages, and data cards.",
    status: "DONE",
    priority: "HIGH",
    estimatedHours: 10,
    deadline: dateInDays(10),
    orderIndex: 1,
    source: "AI",
    editedByUser: true,
    createdAt: "2026-04-14T10:20:00.000Z",
    updatedAt: "2026-04-29T09:10:00.000Z"
  },
  {
    id: "bf50f527-0c3f-4f09-8498-1d93cd88bbff",
    goalId: GOAL_ONE_ID,
    parentTaskId: "fc1118f6-b0dc-43cb-8c9d-46c889474e55",
    title: "Implement typed API client",
    description: "Centralize authentication, goal, task, and analytics requests.",
    status: "IN_PROGRESS",
    priority: "HIGH",
    estimatedHours: 8,
    deadline: dateInDays(12),
    orderIndex: 2,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-14T10:40:00.000Z",
    updatedAt: "2026-05-01T12:30:00.000Z"
  },
  {
    id: "dcb62262-cc3b-4f0a-9fcf-76ceaf88a09d",
    goalId: GOAL_ONE_ID,
    parentTaskId: "bf50f527-0c3f-4f09-8498-1d93cd88bbff",
    title: "Handle JWT lifecycle and 401 redirects",
    description: "Ensure protected routes are stable and token resets are automatic.",
    status: "TODO",
    priority: "HIGH",
    estimatedHours: 3,
    deadline: dateInDays(14),
    orderIndex: 1,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-14T11:00:00.000Z",
    updatedAt: "2026-04-14T11:00:00.000Z"
  },
  {
    id: "5db4756c-61b4-4b35-9368-39ec6d5fe5e7",
    goalId: GOAL_ONE_ID,
    parentTaskId: "bf50f527-0c3f-4f09-8498-1d93cd88bbff",
    title: "Support frontend-only mock mode",
    description: "Provide realistic demo data without creating a local backend.",
    status: "TODO",
    priority: "MEDIUM",
    estimatedHours: 4,
    deadline: dateInDays(15),
    orderIndex: 2,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-14T11:05:00.000Z",
    updatedAt: "2026-04-14T11:05:00.000Z"
  },
  {
    id: "cfa766e4-7fc9-4496-bfb5-b8d4c3620953",
    goalId: GOAL_ONE_ID,
    parentTaskId: null,
    title: "Stage 3. Review and defense readiness",
    description: "Verify plan quality, feasibility, and presentation coherence.",
    status: "TODO",
    priority: "MEDIUM",
    estimatedHours: 14,
    deadline: dateInDays(30),
    orderIndex: 3,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-15T09:00:00.000Z",
    updatedAt: "2026-04-15T09:00:00.000Z"
  },
  {
    id: "9f0a1290-c42c-4955-96b8-8b93b9bf3fda",
    goalId: GOAL_ONE_ID,
    parentTaskId: "cfa766e4-7fc9-4496-bfb5-b8d4c3620953",
    title: "Review feasibility assumptions",
    description: "Compare the required hours with the time available until the deadline.",
    status: "TODO",
    priority: "MEDIUM",
    estimatedHours: 5,
    deadline: dateInDays(24),
    orderIndex: 1,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-15T09:20:00.000Z",
    updatedAt: "2026-04-15T09:20:00.000Z"
  },
  {
    id: "837d8631-2556-4c68-a410-b5f8994cecc6",
    goalId: GOAL_ONE_ID,
    parentTaskId: "cfa766e4-7fc9-4496-bfb5-b8d4c3620953",
    title: "Prepare demo narrative",
    description: "Connect the user flow, the AI plan, and the project results into a clear story.",
    status: "TODO",
    priority: "MEDIUM",
    estimatedHours: 4,
    deadline: dateInDays(26),
    orderIndex: 2,
    source: "AI",
    editedByUser: false,
    createdAt: "2026-04-15T09:35:00.000Z",
    updatedAt: "2026-04-15T09:35:00.000Z"
  }
];

let taskDependencies: TaskDependency[] = [
  {
    id: "8dd5bdbc-a095-4065-946a-e8f85650d802",
    taskId: "bf50f527-0c3f-4f09-8498-1d93cd88bbff",
    dependsOnTaskId: "a7516263-5512-4ab3-b377-3644f8640f97",
    createdAt: "2026-04-18T10:20:00.000Z"
  },
  {
    id: "4f7277fc-1ed0-45f6-8a18-6b1f58f2238d",
    taskId: "9f0a1290-c42c-4955-96b8-8b93b9bf3fda",
    dependsOnTaskId: "bf50f527-0c3f-4f09-8498-1d93cd88bbff",
    createdAt: "2026-04-22T08:45:00.000Z"
  }
];

let planVersions: PlanVersion[] = [
  {
    id: "4b5d1d17-7a63-4965-b2fc-b7a983b3e522",
    goalId: GOAL_ONE_ID,
    generationId: "eb90a95c-4071-43d5-9c86-4f4704689182",
    versionNumber: 1,
    isActive: false,
    createdAt: "2026-04-13T09:00:00.000Z"
  },
  {
    id: "cf1d3b62-92ef-41ab-9d6a-7c2c0c7025d0",
    goalId: GOAL_ONE_ID,
    generationId: "33767b3e-c8a6-4a9c-9160-6d1ce50de9ad",
    versionNumber: 2,
    isActive: true,
    createdAt: "2026-04-24T10:30:00.000Z"
  }
];

let generations: Generation[] = [
  {
    id: "eb90a95c-4071-43d5-9c86-4f4704689182",
    goalId: GOAL_ONE_ID,
    status: "SUCCESS",
    modelName: "gpt-4.1",
    errorMessage: null,
    createdAt: "2026-04-13T09:00:00.000Z"
  },
  {
    id: "33767b3e-c8a6-4a9c-9160-6d1ce50de9ad",
    goalId: GOAL_ONE_ID,
    status: "SUCCESS",
    modelName: "gpt-4.1",
    errorMessage: null,
    createdAt: "2026-04-24T10:30:00.000Z"
  }
];

let generationDetails: Record<string, GenerationDetails> = {
  "eb90a95c-4071-43d5-9c86-4f4704689182": {
    id: "eb90a95c-4071-43d5-9c86-4f4704689182",
    goalId: GOAL_ONE_ID,
    status: "SUCCESS",
    modelName: "gpt-4.1",
    errorMessage: null,
    createdAt: "2026-04-13T09:00:00.000Z",
    prompt:
      "Break down the GoalMind MVP into stages, tasks, and subtasks with priorities and estimated effort.",
    rawAiResponse:
      "{\n  \"plan\": [\n    { \"stage\": \"Product framing\", \"tasks\": 2 },\n    { \"stage\": \"Frontend implementation\", \"tasks\": 5 },\n    { \"stage\": \"Review and defense readiness\", \"tasks\": 2 }\n  ]\n}"
  },
  "33767b3e-c8a6-4a9c-9160-6d1ce50de9ad": {
    id: "33767b3e-c8a6-4a9c-9160-6d1ce50de9ad",
    goalId: GOAL_ONE_ID,
    status: "SUCCESS",
    modelName: "gpt-4.1",
    errorMessage: null,
    createdAt: "2026-04-24T10:30:00.000Z",
    prompt:
      "Refine the GoalMind MVP plan with dependencies, analytics checks, and a stronger defense narrative.",
    rawAiResponse:
      "{\n  \"version\": 2,\n  \"adjustments\": [\"Added dependency review\", \"Expanded analytics coverage\", \"Improved presentation preparation\"]\n}"
  }
};

let feedbackEntries: Feedback[] = [];

let goalAuditMetrics: Record<string, GoalAuditMetrics> = {
  [GOAL_ONE_ID]: {
    tasksDeletedByUser: 1,
    regenerationCount: 1
  },
  [GOAL_TWO_ID]: {
    tasksDeletedByUser: 0,
    regenerationCount: 0
  }
};

function ensureGoal(goalId: string): Goal {
  const goal = goals.find((item) => item.id === goalId);
  if (!goal) {
    raise(404, "goal_not_found", "The requested goal was not found.");
  }

  return goal;
}

function ensureTask(taskId: string): Task {
  const task = tasks.find((item) => item.id === taskId);
  if (!task) {
    raise(404, "task_not_found", "The requested task was not found.");
  }

  return task;
}

function ensureGeneration(generationId: string): GenerationDetails {
  const generation = generationDetails[generationId];
  if (!generation) {
    raise(404, "generation_not_found", "The requested generation was not found.");
  }

  return generation;
}

function ensureContext(goalId: string): GoalContext {
  const context = goalContexts[goalId];
  if (!context) {
    raise(404, "goal_context_not_found", "Goal context does not exist yet.");
  }

  return context;
}

function blankContext(goalId: string): GoalContext {
  const timestamp = nowIso();
  return {
    id: createUuid(),
    goalId,
    currentLevel: "",
    constraintsText: "",
    preferences: [],
    expectedResult: "",
    createdAt: timestamp,
    updatedAt: timestamp
  };
}

function touchGoal(goalId: string) {
  const goal = ensureGoal(goalId);
  goal.updatedAt = nowIso();
}

function buildTaskTree(goalId: string): TaskTreeNode[] {
  const goalTasks = tasks
    .filter((item) => item.goalId === goalId)
    .sort(
      (left, right) =>
        left.orderIndex - right.orderIndex ||
        left.createdAt.localeCompare(right.createdAt)
    );

  const byParent = new Map<string | null, Task[]>();
  goalTasks.forEach((task) => {
    const siblings = byParent.get(task.parentTaskId) || [];
    siblings.push(task);
    byParent.set(task.parentTaskId, siblings);
  });

  const assemble = (parentTaskId: string | null): TaskTreeNode[] =>
    (byParent.get(parentTaskId) || [])
      .sort(
        (left, right) =>
          left.orderIndex - right.orderIndex ||
          left.createdAt.localeCompare(right.createdAt)
      )
      .map((task) => ({
        ...task,
        children: assemble(task.id)
      }));

  return assemble(null);
}

function findDescendantIds(taskId: string): string[] {
  const children = tasks.filter((task) => task.parentTaskId === taskId);
  return children.flatMap((task) => [task.id, ...findDescendantIds(task.id)]);
}

function computeProgress(goalId: string): ProgressResponse {
  const goalTasks = tasks.filter((task) => task.goalId === goalId);
  const completed = goalTasks.filter((task) => task.status === "DONE");
  const totalEstimatedHours = goalTasks.reduce(
    (sum, task) => sum + task.estimatedHours,
    0
  );
  const completedEstimatedHours = completed.reduce(
    (sum, task) => sum + task.estimatedHours,
    0
  );

  return {
    goalId,
    totalTasks: goalTasks.length,
    completedTasks: completed.length,
    totalEstimatedHours,
    completedEstimatedHours,
    simpleProgressPercent:
      goalTasks.length > 0 ? Math.round((completed.length / goalTasks.length) * 100) : 0,
    weightedProgressPercent:
      totalEstimatedHours > 0
        ? Math.round((completedEstimatedHours / totalEstimatedHours) * 100)
        : 0
  };
}

function computeFeasibility(goalId: string): FeasibilityResponse {
  const goal = ensureGoal(goalId);
  const progress = computeProgress(goalId);

  if (!goal.deadline) {
    return {
      goalId,
      weeksUntilDeadline: 0,
      availableHours: 0,
      requiredHours: Math.max(0, progress.totalEstimatedHours - progress.completedEstimatedHours),
      status: "UNKNOWN",
      message: "Set a deadline to calculate whether the remaining plan fits your schedule."
    };
  }

  const now = Date.now();
  const deadlineTime = new Date(goal.deadline).getTime();
  const weeksUntilDeadline = Math.max(0, Math.ceil((deadlineTime - now) / (7 * DAY_MS)));
  const availableHours = weeksUntilDeadline * goal.availableHoursPerWeek;
  const requiredHours = Math.max(
    0,
    progress.totalEstimatedHours - progress.completedEstimatedHours
  );

  let status: FeasibilityResponse["status"] = "UNKNOWN";
  let message =
    "There is not enough data yet to confidently evaluate feasibility.";

  if (requiredHours === 0) {
    status = "REALISTIC";
    message = "All estimated work is already completed.";
  } else if (availableHours === 0 && requiredHours > 0) {
    status = "UNREALISTIC";
    message = "The deadline leaves no remaining weekly capacity for unfinished work.";
  } else if (requiredHours <= availableHours * 0.8) {
    status = "REALISTIC";
    message = "The plan fits comfortably into the available weekly capacity.";
  } else if (requiredHours <= availableHours * 1.1) {
    status = "RISKY";
    message = "The plan is close to capacity limits and may require tighter execution.";
  } else {
    status = "UNREALISTIC";
    message = "The estimated effort exceeds the hours available before the deadline.";
  }

  return {
    goalId,
    weeksUntilDeadline,
    availableHours,
    requiredHours,
    status,
    message
  };
}

function computeQuality(goalId: string): QualityResponse {
  const goal = ensureGoal(goalId);
  const goalTasks = tasks.filter((task) => task.goalId === goalId);
  const rootTasks = goalTasks.filter((task) => task.parentTaskId === null);
  const dependencyCount = taskDependencies.filter((dependency) => {
    const task = tasks.find((item) => item.id === dependency.taskId);
    return task?.goalId === goalId;
  }).length;
  const context = goalContexts[goalId];
  const hasContextDepth =
    Boolean(context?.currentLevel) &&
    Boolean(context?.constraintsText) &&
    context.preferences.length > 0 &&
    Boolean(context.expectedResult);

  const score = Math.min(
    100,
    38 +
      Math.min(goalTasks.length, 12) * 4 +
      Math.min(rootTasks.length, 4) * 5 +
      dependencyCount * 4 +
      (goal.deadline ? 5 : 0) +
      (hasContextDepth ? 10 : 0)
  );

  const details = [
    rootTasks.length >= 3
      ? "The plan is grouped into clear stages."
      : "Add more top-level stages to improve structural readability.",
    dependencyCount > 0
      ? "Dependencies capture execution order between key tasks."
      : "No dependencies are defined yet; consider sequencing critical work.",
    hasContextDepth
      ? "Goal context is detailed enough to support stronger AI decomposition."
      : "Goal context can be expanded with clearer constraints and preferences."
  ];

  return {
    goalId,
    score,
    details
  };
}

function computeMetrics(goalId: string): MetricsResponse {
  const goalTasks = tasks.filter((task) => task.goalId === goalId);
  const completedTasks = goalTasks.filter((task) => task.status === "DONE").length;
  const acceptedTasks = goalTasks.filter((task) => task.status !== "CANCELLED").length;
  const audit = goalAuditMetrics[goalId] || {
    tasksDeletedByUser: 0,
    regenerationCount: 0
  };

  return {
    goalId,
    tasksCreatedByAI: goalTasks.filter((task) => task.source === "AI").length,
    tasksEditedByUser: goalTasks.filter((task) => task.editedByUser).length,
    tasksDeletedByUser: audit.tasksDeletedByUser,
    tasksCompleted: completedTasks,
    regenerationCount: audit.regenerationCount,
    acceptedTasksPercent:
      goalTasks.length > 0 ? Math.round((acceptedTasks / goalTasks.length) * 100) : 0
  };
}

function nextSiblingOrder(goalId: string, parentTaskId: string | null) {
  const siblings = tasks.filter(
    (task) => task.goalId === goalId && task.parentTaskId === parentTaskId
  );
  return siblings.length === 0
    ? 1
    : Math.max(...siblings.map((task) => task.orderIndex)) + 1;
}

function createTaskRecord(input: {
  goalId: string;
  parentTaskId: string | null;
  title: string;
  description: string;
  priority: Task["priority"];
  estimatedHours: number;
  orderIndex: number;
  deadline: string | null;
  source: Task["source"];
  editedByUser: boolean;
}): Task {
  const timestamp = nowIso();
  return {
    id: createUuid(),
    goalId: input.goalId,
    parentTaskId: input.parentTaskId,
    title: input.title,
    description: input.description,
    status: "TODO",
    priority: input.priority,
    estimatedHours: input.estimatedHours,
    deadline: input.deadline,
    orderIndex: input.orderIndex,
    source: input.source,
    editedByUser: input.editedByUser,
    createdAt: timestamp,
    updatedAt: timestamp
  };
}

function generateDeadline(goal: Goal, ratio: number, enabled: boolean) {
  if (!enabled || !goal.deadline) {
    return null;
  }

  const deadlineTime = new Date(goal.deadline).getTime();
  const remainingDays = Math.max(
    3,
    Math.ceil((deadlineTime - Date.now()) / DAY_MS)
  );
  const suggestedDays = Math.max(2, Math.floor(remainingDays * ratio));
  return dateInDays(suggestedDays);
}

function createGeneratedTasks(goal: Goal, request: DecomposeGoalRequest) {
  const items: Task[] = [];
  const stagePlanning = createTaskRecord({
    goalId: goal.id,
    parentTaskId: null,
    title: "Stage 1. Frame the outcome",
    description: "Clarify what success looks like, who the work serves, and how the result will be evaluated.",
    priority: "HIGH",
    estimatedHours: request.detailLevel === "LOW" ? 5 : 7,
    orderIndex: nextSiblingOrder(goal.id, null),
    deadline: generateDeadline(goal, 0.2, request.includeDeadlines),
    source: "AI",
    editedByUser: false
  });
  items.push(stagePlanning);

  items.push(
    createTaskRecord({
      goalId: goal.id,
      parentTaskId: stagePlanning.id,
      title: `Define success criteria for "${goal.title}"`,
      description: "Translate the high-level goal into measurable completion signals.",
      priority: "HIGH",
      estimatedHours: 3,
      orderIndex: 1,
      deadline: generateDeadline(goal, 0.16, request.includeDeadlines),
      source: "AI",
      editedByUser: false
    })
  );

  items.push(
    createTaskRecord({
      goalId: goal.id,
      parentTaskId: stagePlanning.id,
      title: "Map constraints and resources",
      description: "Capture limitations, tools, collaborators, and available weekly capacity.",
      priority: "MEDIUM",
      estimatedHours: 3,
      orderIndex: 2,
      deadline: generateDeadline(goal, 0.2, request.includeDeadlines),
      source: "AI",
      editedByUser: false
    })
  );

  const stageExecution = createTaskRecord({
    goalId: goal.id,
    parentTaskId: null,
    title: "Stage 2. Build the execution plan",
    description: "Turn the goal into milestones, deliverables, and concrete implementation work.",
    priority: "HIGH",
    estimatedHours: request.detailLevel === "HIGH" ? 12 : 9,
    orderIndex: nextSiblingOrder(goal.id, null) + 1,
    deadline: generateDeadline(goal, 0.55, request.includeDeadlines),
    source: "AI",
    editedByUser: false
  });
  items.push(stageExecution);

  const milestoneTask = createTaskRecord({
    goalId: goal.id,
    parentTaskId: stageExecution.id,
    title: "Create milestone roadmap",
    description: "Split the work into sequential milestones with clear ownership and outcomes.",
    priority: "HIGH",
    estimatedHours: 5,
    orderIndex: 1,
    deadline: generateDeadline(goal, 0.38, request.includeDeadlines),
    source: "AI",
    editedByUser: false
  });
  items.push(milestoneTask);

  items.push(
    createTaskRecord({
      goalId: goal.id,
      parentTaskId: milestoneTask.id,
      title: "Prepare milestone checklist",
      description: "List the deliverables that indicate milestone completion.",
      priority: "MEDIUM",
      estimatedHours: 2,
      orderIndex: 1,
      deadline: generateDeadline(goal, 0.34, request.includeDeadlines),
      source: "AI",
      editedByUser: false
    })
  );

  items.push(
    createTaskRecord({
      goalId: goal.id,
      parentTaskId: milestoneTask.id,
      title: "Estimate effort for each milestone",
      description: "Assign rough hours to the milestone sequence for feasibility checks.",
      priority: "MEDIUM",
      estimatedHours: 2,
      orderIndex: 2,
      deadline: generateDeadline(goal, 0.4, request.includeDeadlines),
      source: "AI",
      editedByUser: false
    })
  );

  items.push(
    createTaskRecord({
      goalId: goal.id,
      parentTaskId: stageExecution.id,
      title: "Draft dependency map",
      description: "Identify which tasks block downstream work and need early attention.",
      priority: request.includeDependencies ? "HIGH" : "MEDIUM",
      estimatedHours: request.includeDependencies ? 4 : 2,
      orderIndex: 2,
      deadline: generateDeadline(goal, 0.48, request.includeDeadlines),
      source: "AI",
      editedByUser: false
    })
  );

  const stageValidation = createTaskRecord({
    goalId: goal.id,
    parentTaskId: null,
    title: "Stage 3. Review and prove feasibility",
    description: "Assess whether the plan is realistic and prepare the final demonstration flow.",
    priority: "MEDIUM",
    estimatedHours: request.detailLevel === "HIGH" ? 8 : 6,
    orderIndex: nextSiblingOrder(goal.id, null) + 2,
    deadline: generateDeadline(goal, 0.82, request.includeDeadlines),
    source: "AI",
    editedByUser: false
  });
  items.push(stageValidation);

  items.push(
    createTaskRecord({
      goalId: goal.id,
      parentTaskId: stageValidation.id,
      title: "Review feasibility and buffer time",
      description: "Compare remaining estimated work against available weekly hours.",
      priority: "MEDIUM",
      estimatedHours: 3,
      orderIndex: 1,
      deadline: generateDeadline(goal, 0.72, request.includeDeadlines),
      source: "AI",
      editedByUser: false
    })
  );

  items.push(
    createTaskRecord({
      goalId: goal.id,
      parentTaskId: stageValidation.id,
      title: "Prepare final walkthrough",
      description: "Organize the demonstration flow and decision points for the completed plan.",
      priority: "MEDIUM",
      estimatedHours: 3,
      orderIndex: 2,
      deadline: generateDeadline(goal, 0.8, request.includeDeadlines),
      source: "AI",
      editedByUser: false
    })
  );

  if (request.detailLevel === "HIGH") {
    items.push(
      createTaskRecord({
        goalId: goal.id,
        parentTaskId: stageValidation.id,
        title: "Document fallback strategy",
        description: "Prepare mitigation steps for risks that could block the final delivery.",
        priority: "MEDIUM",
        estimatedHours: 2,
        orderIndex: 3,
        deadline: generateDeadline(goal, 0.78, request.includeDeadlines),
        source: "AI",
        editedByUser: false
      })
    );
  }

  return items;
}

function addGeneratedDependencies(goalId: string, createdTasks: Task[]) {
  if (createdTasks.length < 4) {
    return;
  }

  const taskByTitle = new Map(createdTasks.map((task) => [task.title, task]));
  const milestoneTask = taskByTitle.get("Create milestone roadmap");
  const dependencyMapTask = taskByTitle.get("Draft dependency map");
  const feasibilityTask = taskByTitle.get("Review feasibility and buffer time");

  const newDependencies: TaskDependency[] = [];
  if (dependencyMapTask && milestoneTask) {
    newDependencies.push({
      id: createUuid(),
      taskId: dependencyMapTask.id,
      dependsOnTaskId: milestoneTask.id,
      createdAt: nowIso()
    });
  }

  if (feasibilityTask && dependencyMapTask) {
    newDependencies.push({
      id: createUuid(),
      taskId: feasibilityTask.id,
      dependsOnTaskId: dependencyMapTask.id,
      createdAt: nowIso()
    });
  }

  if (newDependencies.length > 0) {
    taskDependencies = [...taskDependencies, ...newDependencies];
    touchGoal(goalId);
  }
}

function appendRegenerationRefinement(goal: Goal, request: DecomposeGoalRequest) {
  const refinementStage = createTaskRecord({
    goalId: goal.id,
    parentTaskId: null,
    title: `Stage ${tasks.filter((task) => task.goalId === goal.id && task.parentTaskId === null).length + 1}. Refinement pass`,
    description: "Capture improvements from a fresh AI generation without removing previous work.",
    priority: "MEDIUM",
    estimatedHours: request.detailLevel === "HIGH" ? 6 : 4,
    orderIndex: nextSiblingOrder(goal.id, null),
    deadline: generateDeadline(goal, 0.88, request.includeDeadlines),
    source: "AI",
    editedByUser: false
  });

  const reviewTask = createTaskRecord({
    goalId: goal.id,
    parentTaskId: refinementStage.id,
    title: "Review plan gaps and overlaps",
    description: "Identify duplicated work, missing checkpoints, and under-specified dependencies.",
    priority: "MEDIUM",
    estimatedHours: 2,
    orderIndex: 1,
    deadline: generateDeadline(goal, 0.84, request.includeDeadlines),
    source: "AI",
    editedByUser: false
  });

  const adjustmentTask = createTaskRecord({
    goalId: goal.id,
    parentTaskId: refinementStage.id,
    title: "Adjust sequence for stronger delivery confidence",
    description: "Re-order tasks to reduce risk before the final deadline window.",
    priority: "MEDIUM",
    estimatedHours: 2,
    orderIndex: 2,
    deadline: generateDeadline(goal, 0.87, request.includeDeadlines),
    source: "AI",
    editedByUser: false
  });

  return [refinementStage, reviewTask, adjustmentTask];
}

function createGeneration(goalId: string, prompt: string, rawAiResponse: string) {
  const timestamp = nowIso();
  const generation: Generation = {
    id: createUuid(),
    goalId,
    status: "SUCCESS",
    modelName: "gpt-4.1",
    errorMessage: null,
    createdAt: timestamp
  };

  const details: GenerationDetails = {
    ...generation,
    prompt,
    rawAiResponse
  };

  generations = [generation, ...generations];
  generationDetails[generation.id] = details;

  return generation;
}

function createPlanVersion(goalId: string, generationId: string) {
  const currentVersions = planVersions.filter((version) => version.goalId === goalId);
  const versionNumber =
    currentVersions.length > 0
      ? Math.max(...currentVersions.map((version) => version.versionNumber)) + 1
      : 1;

  planVersions = planVersions.map((version) =>
    version.goalId === goalId ? { ...version, isActive: false } : version
  );

  const version: PlanVersion = {
    id: createUuid(),
    goalId,
    generationId,
    versionNumber,
    isActive: true,
    createdAt: nowIso()
  };

  planVersions = [version, ...planVersions];
  return version;
}

function runGoalDecomposition(
  goalId: string,
  request: DecomposeGoalRequest,
  mode: "initial" | "regenerate"
): DecomposeGoalResponse {
  const goal = ensureGoal(goalId);
  const existingTaskCount = tasks.filter((task) => task.goalId === goalId).length;
  const createdTasks =
    existingTaskCount === 0 || mode === "initial"
      ? createGeneratedTasks(goal, request)
      : appendRegenerationRefinement(goal, request);

  tasks = [...tasks, ...createdTasks];

  if (request.includeDependencies) {
    addGeneratedDependencies(goalId, createdTasks);
  }

  const prompt =
    mode === "initial"
      ? `Decompose the goal "${goal.title}" into a structured execution plan with ${request.detailLevel.toLowerCase()} detail.`
      : `Regenerate and refine the goal "${goal.title}" with updated structure, dependencies, and feasibility checks.`;
  const rawAiResponse = JSON.stringify(
    {
      goalId,
      mode,
      tasksCreated: createdTasks.length,
      tasks: createdTasks.map((task) => ({
        title: task.title,
        parentTaskId: task.parentTaskId,
        estimatedHours: task.estimatedHours,
        priority: task.priority
      }))
    },
    null,
    2
  );

  const generation = createGeneration(goalId, prompt, rawAiResponse);
  const planVersion = createPlanVersion(goalId, generation.id);
  goal.status = "ACTIVE";
  goal.updatedAt = nowIso();

  if (mode === "regenerate") {
    const audit = goalAuditMetrics[goalId] || {
      tasksDeletedByUser: 0,
      regenerationCount: 0
    };
    audit.regenerationCount += 1;
    goalAuditMetrics[goalId] = audit;
  }

  const quality = computeQuality(goalId);
  const feasibility = computeFeasibility(goalId);

  return {
    goalId,
    generationId: generation.id,
    planVersionId: planVersion.id,
    tasksCreated: createdTasks.length,
    qualityScore: quality.score,
    feasibilityStatus: feasibility.status
  };
}

export const mockApi = {
  demoToken: DEMO_TOKEN,

  login(payload: LoginRequest): Promise<AuthResponse> {
    return respond(() => {
      if (!payload.email || !payload.password) {
        raise(400, "validation_error", "Email and password are required.");
      }

      currentUser = {
        ...currentUser,
        email: payload.email
      };

      return {
        accessToken: DEMO_TOKEN,
        user: currentUser
      };
    });
  },

  register(payload: RegisterRequest): Promise<AuthResponse> {
    return respond(() => {
      if (!payload.name || !payload.email || !payload.password) {
        raise(400, "validation_error", "Name, email, and password are required.");
      }

      currentUser = {
        ...currentUser,
        name: payload.name,
        email: payload.email
      };

      return {
        accessToken: DEMO_TOKEN,
        user: currentUser
      };
    });
  },

  getMe(): Promise<User> {
    return respond(() => currentUser, 180);
  },

  getAISettings(): Promise<AISettingsResponse> {
    return respond(() => mockAISettings, 180);
  },

  saveAISettings(payload: AISettingsRequest): Promise<AISettingsResponse> {
    return respond(() => {
      if (!payload.provider || !payload.apiKey.trim()) {
        raise(400, "validation_error", "Provider and API key are required.");
      }

      const key = payload.apiKey.trim();
      mockAISettings = {
        provider: payload.provider,
        maskedKey:
          key.length <= 8
            ? `${key.slice(0, 2)}***`
            : `${key.slice(0, 4)}...${key.slice(-4)}`
      };
      return mockAISettings;
    }, 240);
  },

  listGoals(): Promise<Goal[]> {
    return respond(() =>
      [...goals].sort((left, right) => right.updatedAt.localeCompare(left.updatedAt))
    );
  },

  createGoal(payload: CreateGoalRequest): Promise<Goal> {
    return respond(() => {
      if (!payload.title.trim()) {
        raise(400, "validation_error", "Goal title is required.");
      }

      const timestamp = nowIso();
      const goal: Goal = {
        id: createUuid(),
        title: payload.title.trim(),
        description: payload.description.trim(),
        category: payload.category.trim() || "General",
        priority: payload.priority,
        status: "DRAFT",
        deadline: payload.deadline || null,
        availableHoursPerWeek: payload.availableHoursPerWeek,
        createdAt: timestamp,
        updatedAt: timestamp
      };

      goals = [goal, ...goals];
      goalContexts[goal.id] = blankContext(goal.id);
      goalAuditMetrics[goal.id] = {
        tasksDeletedByUser: 0,
        regenerationCount: 0
      };

      return goal;
    });
  },

  getGoal(goalId: string): Promise<Goal> {
    return respond(() => ensureGoal(goalId));
  },

  updateGoal(goalId: string, payload: UpdateGoalRequest): Promise<Goal> {
    return respond(() => {
      const goal = ensureGoal(goalId);
      Object.assign(goal, {
        ...payload,
        deadline: payload.deadline === undefined ? goal.deadline : payload.deadline,
        updatedAt: nowIso()
      });
      return goal;
    });
  },

  deleteGoal(goalId: string): Promise<void> {
    return respond(() => {
      ensureGoal(goalId);

      const taskIds = tasks.filter((task) => task.goalId === goalId).map((task) => task.id);
      goals = goals.filter((goal) => goal.id !== goalId);
      tasks = tasks.filter((task) => task.goalId !== goalId);
      taskDependencies = taskDependencies.filter(
        (dependency) =>
          !taskIds.includes(dependency.taskId) &&
          !taskIds.includes(dependency.dependsOnTaskId)
      );
      planVersions = planVersions.filter((version) => version.goalId !== goalId);
      generations = generations.filter((generation) => generation.goalId !== goalId);
      Object.keys(generationDetails).forEach((generationId) => {
        if (generationDetails[generationId].goalId === goalId) {
          delete generationDetails[generationId];
        }
      });
      feedbackEntries = feedbackEntries.filter((feedback) => {
        const details = generationDetails[feedback.generationId];
        return details ? details.goalId !== goalId : false;
      });
      delete goalContexts[goalId];
      delete goalAuditMetrics[goalId];
    });
  },

  getGoalContext(goalId: string): Promise<GoalContext> {
    return respond(() => ensureContext(goalId));
  },

  createGoalContext(goalId: string, payload: GoalContextInput): Promise<GoalContext> {
    return respond(() => {
      ensureGoal(goalId);
      if (goalContexts[goalId]) {
        raise(409, "goal_context_exists", "Goal context already exists.");
      }

      const timestamp = nowIso();
      const context: GoalContext = {
        id: createUuid(),
        goalId,
        currentLevel: payload.currentLevel,
        constraintsText: payload.constraintsText,
        preferences: payload.preferences,
        expectedResult: payload.expectedResult,
        createdAt: timestamp,
        updatedAt: timestamp
      };
      goalContexts[goalId] = context;
      touchGoal(goalId);
      return context;
    });
  },

  updateGoalContext(goalId: string, payload: GoalContextInput): Promise<GoalContext> {
    return respond(() => {
      const context = ensureContext(goalId);
      Object.assign(context, {
        ...payload,
        updatedAt: nowIso()
      });
      touchGoal(goalId);
      return context;
    });
  },

  listGoalTasks(goalId: string): Promise<TaskTreeNode[]> {
    return respond(() => {
      ensureGoal(goalId);
      return buildTaskTree(goalId);
    });
  },

  createTask(goalId: string, payload: CreateTaskRequest): Promise<Task> {
    return respond(() => {
      ensureGoal(goalId);
      if (payload.parentTaskId) {
        ensureTask(payload.parentTaskId);
      }

      const task = createTaskRecord({
        goalId,
        parentTaskId: payload.parentTaskId ?? null,
        title: payload.title.trim(),
        description: payload.description.trim(),
        priority: payload.priority,
        estimatedHours: payload.estimatedHours,
        orderIndex:
          payload.orderIndex ?? nextSiblingOrder(goalId, payload.parentTaskId ?? null),
        deadline: payload.deadline ?? null,
        source: "MANUAL",
        editedByUser: true
      });
      task.status = payload.status ?? "TODO";
      tasks = [...tasks, task];
      touchGoal(goalId);
      return task;
    });
  },

  getTask(taskId: string): Promise<Task> {
    return respond(() => ensureTask(taskId));
  },

  updateTask(taskId: string, payload: UpdateTaskRequest): Promise<Task> {
    return respond(() => {
      const task = ensureTask(taskId);
      Object.assign(task, {
        ...payload,
        editedByUser: payload.editedByUser ?? true,
        updatedAt: nowIso()
      });
      touchGoal(task.goalId);
      return task;
    });
  },

  deleteTask(taskId: string): Promise<void> {
    return respond(() => {
      const task = ensureTask(taskId);
      const idsToDelete = [task.id, ...findDescendantIds(task.id)];
      tasks = tasks.filter((item) => !idsToDelete.includes(item.id));
      taskDependencies = taskDependencies.filter(
        (dependency) =>
          !idsToDelete.includes(dependency.taskId) &&
          !idsToDelete.includes(dependency.dependsOnTaskId)
      );
      const audit = goalAuditMetrics[task.goalId] || {
        tasksDeletedByUser: 0,
        regenerationCount: 0
      };
      audit.tasksDeletedByUser += idsToDelete.length;
      goalAuditMetrics[task.goalId] = audit;
      touchGoal(task.goalId);
    });
  },

  updateTaskStatus(
    taskId: string,
    payload: UpdateTaskStatusRequest
  ): Promise<Task> {
    return respond(() => {
      const task = ensureTask(taskId);
      task.status = payload.status;
      task.editedByUser = true;
      task.updatedAt = nowIso();
      touchGoal(task.goalId);
      return task;
    });
  },

  listDependencies(taskId: string): Promise<TaskDependency[]> {
    return respond(() => {
      ensureTask(taskId);
      return taskDependencies.filter((dependency) => dependency.taskId === taskId);
    });
  },

  createDependency(
    taskId: string,
    payload: CreateTaskDependencyRequest
  ): Promise<TaskDependency> {
    return respond(() => {
      const task = ensureTask(taskId);
      const dependencyTarget = ensureTask(payload.dependsOnTaskId);

      if (task.id === dependencyTarget.id) {
        raise(409, "dependency_invalid", "A task cannot depend on itself.");
      }

      if (task.goalId !== dependencyTarget.goalId) {
        raise(
          409,
          "dependency_cross_goal",
          "Dependencies must reference tasks from the same goal."
        );
      }

      const exists = taskDependencies.some(
        (dependency) =>
          dependency.taskId === taskId &&
          dependency.dependsOnTaskId === payload.dependsOnTaskId
      );

      if (exists) {
        raise(409, "dependency_exists", "This dependency already exists.");
      }

      const dependency: TaskDependency = {
        id: createUuid(),
        taskId,
        dependsOnTaskId: payload.dependsOnTaskId,
        createdAt: nowIso()
      };
      taskDependencies = [...taskDependencies, dependency];
      touchGoal(task.goalId);
      return dependency;
    });
  },

  deleteDependency(taskId: string, dependencyId: string): Promise<void> {
    return respond(() => {
      const task = ensureTask(taskId);
      const existing = taskDependencies.find(
        (dependency) =>
          dependency.id === dependencyId && dependency.taskId === taskId
      );

      if (!existing) {
        raise(404, "dependency_not_found", "Dependency was not found.");
      }

      taskDependencies = taskDependencies.filter(
        (dependency) => dependency.id !== dependencyId
      );
      touchGoal(task.goalId);
    });
  },

  decomposeGoal(
    goalId: string,
    payload: DecomposeGoalRequest
  ): Promise<DecomposeGoalResponse> {
    return respond(() => runGoalDecomposition(goalId, payload, "initial"), 360);
  },

  regenerateGoal(
    goalId: string,
    payload: DecomposeGoalRequest
  ): Promise<DecomposeGoalResponse> {
    return respond(() => runGoalDecomposition(goalId, payload, "regenerate"), 360);
  },

  decomposeTask(
    taskId: string,
    payload: DecomposeGoalRequest
  ): Promise<DecomposeGoalResponse> {
    return respond(() => {
      const task = ensureTask(taskId);
      const createdTasks = [
        createTaskRecord({
          goalId: task.goalId,
          parentTaskId: task.id,
          title: `Break down: ${task.title}`,
          description: "Create a more detailed implementation or execution checklist.",
          priority: task.priority,
          estimatedHours: 2,
          orderIndex: nextSiblingOrder(task.goalId, task.id),
          deadline: task.deadline,
          source: "AI",
          editedByUser: false
        }),
        createTaskRecord({
          goalId: task.goalId,
          parentTaskId: task.id,
          title: `Verify results for ${task.title}`,
          description: "Add a validation checkpoint so the parent task has a measurable outcome.",
          priority: payload.detailLevel === "HIGH" ? "HIGH" : "MEDIUM",
          estimatedHours: 2,
          orderIndex: nextSiblingOrder(task.goalId, task.id) + 1,
          deadline: task.deadline,
          source: "AI",
          editedByUser: false
        })
      ];

      tasks = [...tasks, ...createdTasks];
      const generation = createGeneration(
        task.goalId,
        `Decompose the task "${task.title}" into more granular actions.`,
        JSON.stringify(
          {
            taskId,
            tasksCreated: createdTasks.length,
            detailLevel: payload.detailLevel
          },
          null,
          2
        )
      );
      const version = createPlanVersion(task.goalId, generation.id);
      touchGoal(task.goalId);

      const quality = computeQuality(task.goalId);
      const feasibility = computeFeasibility(task.goalId);

      return {
        goalId: task.goalId,
        generationId: generation.id,
        planVersionId: version.id,
        tasksCreated: createdTasks.length,
        qualityScore: quality.score,
        feasibilityStatus: feasibility.status
      };
    }, 360);
  },

  getProgress(goalId: string): Promise<ProgressResponse> {
    return respond(() => computeProgress(goalId));
  },

  getFeasibility(goalId: string): Promise<FeasibilityResponse> {
    return respond(() => computeFeasibility(goalId));
  },

  getQuality(goalId: string): Promise<QualityResponse> {
    return respond(() => computeQuality(goalId));
  },

  getMetrics(goalId: string): Promise<MetricsResponse> {
    return respond(() => computeMetrics(goalId));
  },

  listPlanVersions(goalId: string): Promise<PlanVersion[]> {
    return respond(() => {
      ensureGoal(goalId);
      return planVersions
        .filter((version) => version.goalId === goalId)
        .sort((left, right) => right.versionNumber - left.versionNumber);
    });
  },

  getPlanVersion(goalId: string, versionId: string): Promise<PlanVersion> {
    return respond(() => {
      ensureGoal(goalId);
      const version = planVersions.find(
        (item) => item.goalId === goalId && item.id === versionId
      );
      if (!version) {
        raise(404, "plan_version_not_found", "Plan version was not found.");
      }

      return version;
    });
  },

  activatePlanVersion(goalId: string, versionId: string): Promise<PlanVersion> {
    return respond(() => {
      ensureGoal(goalId);
      const target = planVersions.find(
        (version) => version.goalId === goalId && version.id === versionId
      );
      if (!target) {
        raise(404, "plan_version_not_found", "Plan version was not found.");
      }

      planVersions = planVersions.map((version) =>
        version.goalId === goalId
          ? { ...version, isActive: version.id === versionId }
          : version
      );

      touchGoal(goalId);
      const activatedVersion = planVersions.find(
        (version) => version.goalId === goalId && version.id === versionId
      );

      if (!activatedVersion) {
        raise(404, "plan_version_not_found", "Plan version was not found.");
      }

      return activatedVersion;
    });
  },

  listGenerations(goalId: string): Promise<Generation[]> {
    return respond(() => {
      ensureGoal(goalId);
      return generations
        .filter((generation) => generation.goalId === goalId)
        .sort((left, right) => right.createdAt.localeCompare(left.createdAt));
    });
  },

  getGeneration(generationId: string): Promise<GenerationDetails> {
    return respond(() => ensureGeneration(generationId));
  },

  createFeedback(
    generationId: string,
    payload: FeedbackRequest
  ): Promise<Feedback> {
    return respond(() => {
      ensureGeneration(generationId);
      if (payload.rating < 1 || payload.rating > 5) {
        raise(400, "validation_error", "Rating must be between 1 and 5.");
      }

      const feedback: Feedback = {
        id: createUuid(),
        generationId,
        rating: payload.rating,
        comment: payload.comment.trim(),
        createdAt: nowIso()
      };
      feedbackEntries = [feedback, ...feedbackEntries];
      return feedback;
    });
  }
};

import {
  apiRequest,
  buildApiUrl,
  USE_MOCKS
} from "./client";
import { mockApi } from "./mockData";
import type {
  ApiResponse,
  AISettingsRequest,
  AISettingsResponse,
  AnalyzeGoalRequest,
  AnalyzeGoalResponse,
  AuthResponse,
  CreateGoalRequest,
  CreateTaskDependencyRequest,
  CreateTaskRequest,
  DecomposeGoalRequest,
  DecomposeGoalResponse,
  EvaluateGoalRequestResponse,
  Feedback,
  FeedbackRequest,
  FeasibilityResponse,
  Generation,
  GenerationDetails,
  Goal,
  GoalContext,
  GoalContextInput,
  GeneratePlanFromRequestResponse,
  LoginRequest,
  MetricsResponse,
  PlanVersion,
  ProgressResponse,
  QualityResponse,
  RegisterRequest,
  SubmitGoalRequestAnswersRequest,
  SubmitGoalRequestAnswersResponse,
  Task,
  TaskDependency,
  TaskTreeNode,
  UpdateGoalRequest,
  UpdateTaskRequest,
  UpdateTaskStatusRequest,
  User,
  UserProfile,
  UserProfileInput
} from "./types";

function unwrapData<T>(response: ApiResponse<T>) {
  return response.data;
}

export const authApi = {
  register(payload: RegisterRequest) {
    if (USE_MOCKS) {
      return mockApi.register(payload);
    }

    return apiRequest<ApiResponse<AuthResponse>>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  },

  login(payload: LoginRequest) {
    if (USE_MOCKS) {
      return mockApi.login(payload);
    }

    return apiRequest<ApiResponse<AuthResponse>>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  },

  getMe() {
    if (USE_MOCKS) {
      return mockApi.getMe();
    }

    return apiRequest<ApiResponse<User>>("/api/auth/me").then(unwrapData);
  },

  getGoogleStartUrl() {
    return buildApiUrl("/api/auth/oauth/google/start");
  },

  getGoogleCallbackUrl() {
    return buildApiUrl("/api/auth/oauth/google/callback");
  }
};

export const aiSettingsApi = {
  get() {
    if (USE_MOCKS) {
      return mockApi.getAISettings();
    }

    return apiRequest<ApiResponse<AISettingsResponse>>("/api/ai-settings").then(
      unwrapData
    );
  },

  save(payload: AISettingsRequest) {
    if (USE_MOCKS) {
      return mockApi.saveAISettings(payload);
    }

    return apiRequest<ApiResponse<AISettingsResponse>>("/api/ai-settings", {
      method: "PUT",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  }
};

export const profileApi = {
  get() {
    return apiRequest<ApiResponse<UserProfile>>("/api/profile").then(unwrapData);
  },
  save(payload: UserProfileInput) {
    return apiRequest<ApiResponse<UserProfile>>("/api/profile", {
      method: "PUT",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  }
};

export const goalsApi = {
  list() {
    if (USE_MOCKS) {
      return mockApi.listGoals();
    }

    return apiRequest<ApiResponse<Goal[]>>("/api/goals").then(unwrapData);
  },

  create(payload: CreateGoalRequest) {
    if (USE_MOCKS) {
      return mockApi.createGoal(payload);
    }

    return apiRequest<ApiResponse<Goal>>("/api/goals", {
      method: "POST",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  },

  get(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.getGoal(goalId);
    }

    return apiRequest<ApiResponse<Goal>>(`/api/goals/${goalId}`).then(unwrapData);
  },

  update(goalId: string, payload: UpdateGoalRequest) {
    if (USE_MOCKS) {
      return mockApi.updateGoal(goalId, payload);
    }

    return apiRequest<ApiResponse<Goal>>(`/api/goals/${goalId}`, {
      method: "PUT",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  },

  remove(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.deleteGoal(goalId);
    }

    return apiRequest<void>(`/api/goals/${goalId}`, {
      method: "DELETE"
    });
  }
};

export const requestsApi = {
  analyze(payload: AnalyzeGoalRequest) {
    return apiRequest<ApiResponse<AnalyzeGoalResponse>>("/api/requests/analyze", {
      method: "POST",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  },

  submitAnswers(requestId: string, payload: SubmitGoalRequestAnswersRequest) {
    return apiRequest<ApiResponse<SubmitGoalRequestAnswersResponse>>(
      `/api/requests/${requestId}/answers`,
      {
        method: "POST",
        body: JSON.stringify(payload)
      }
    ).then(unwrapData);
  },

  evaluate(requestId: string) {
    return apiRequest<ApiResponse<EvaluateGoalRequestResponse>>(
      `/api/requests/${requestId}/evaluate`,
      {
        method: "POST"
      }
    ).then(unwrapData);
  },

  generatePlan(requestId: string) {
    return apiRequest<ApiResponse<GeneratePlanFromRequestResponse>>(
      `/api/requests/${requestId}/generate-plan`,
      {
        method: "POST"
      }
    ).then(unwrapData);
  }
};

export const goalContextApi = {
  get(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.getGoalContext(goalId);
    }

    return apiRequest<ApiResponse<GoalContext>>(
      `/api/goals/${goalId}/context`
    ).then(unwrapData);
  },

  create(goalId: string, payload: GoalContextInput) {
    if (USE_MOCKS) {
      return mockApi.createGoalContext(goalId, payload);
    }

    return apiRequest<ApiResponse<GoalContext>>(`/api/goals/${goalId}/context`, {
      method: "POST",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  },

  update(goalId: string, payload: GoalContextInput) {
    if (USE_MOCKS) {
      return mockApi.updateGoalContext(goalId, payload);
    }

    return apiRequest<ApiResponse<GoalContext>>(`/api/goals/${goalId}/context`, {
      method: "PUT",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  }
};

export const tasksApi = {
  listByGoal(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.listGoalTasks(goalId);
    }

    return apiRequest<ApiResponse<TaskTreeNode[]>>(
      `/api/goals/${goalId}/tasks`
    ).then(unwrapData);
  },

  create(goalId: string, payload: CreateTaskRequest) {
    if (USE_MOCKS) {
      return mockApi.createTask(goalId, payload);
    }

    return apiRequest<ApiResponse<Task>>(`/api/goals/${goalId}/tasks`, {
      method: "POST",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  },

  get(taskId: string) {
    if (USE_MOCKS) {
      return mockApi.getTask(taskId);
    }

    return apiRequest<ApiResponse<Task>>(`/api/tasks/${taskId}`).then(unwrapData);
  },

  update(taskId: string, payload: UpdateTaskRequest) {
    if (USE_MOCKS) {
      return mockApi.updateTask(taskId, payload);
    }

    return apiRequest<ApiResponse<Task>>(`/api/tasks/${taskId}`, {
      method: "PUT",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  },

  remove(taskId: string) {
    if (USE_MOCKS) {
      return mockApi.deleteTask(taskId);
    }

    return apiRequest<void>(`/api/tasks/${taskId}`, {
      method: "DELETE"
    });
  },

  updateStatus(taskId: string, payload: UpdateTaskStatusRequest) {
    if (USE_MOCKS) {
      return mockApi.updateTaskStatus(taskId, payload);
    }

    return apiRequest<ApiResponse<Task>>(`/api/tasks/${taskId}/status`, {
      method: "PATCH",
      body: JSON.stringify(payload)
    }).then(unwrapData);
  }
};

export const dependenciesApi = {
  list(taskId: string) {
    if (USE_MOCKS) {
      return mockApi.listDependencies(taskId);
    }

    return apiRequest<ApiResponse<TaskDependency[]>>(
      `/api/tasks/${taskId}/dependencies`
    ).then(unwrapData);
  },

  create(taskId: string, payload: CreateTaskDependencyRequest) {
    if (USE_MOCKS) {
      return mockApi.createDependency(taskId, payload);
    }

    return apiRequest<ApiResponse<TaskDependency>>(
      `/api/tasks/${taskId}/dependencies`,
      {
        method: "POST",
        body: JSON.stringify(payload)
      }
    ).then(unwrapData);
  },

  remove(taskId: string, dependencyId: string) {
    if (USE_MOCKS) {
      return mockApi.deleteDependency(taskId, dependencyId);
    }

    return apiRequest<void>(`/api/tasks/${taskId}/dependencies/${dependencyId}`, {
      method: "DELETE"
    });
  }
};

export const decompositionApi = {
  decomposeGoal(goalId: string, payload: DecomposeGoalRequest) {
    if (USE_MOCKS) {
      return mockApi.decomposeGoal(goalId, payload);
    }

    return apiRequest<ApiResponse<DecomposeGoalResponse>>(
      `/api/goals/${goalId}/decompose`,
      {
        method: "POST",
        body: JSON.stringify(payload)
      }
    ).then(unwrapData);
  },

  regenerateGoal(goalId: string, payload: DecomposeGoalRequest) {
    if (USE_MOCKS) {
      return mockApi.regenerateGoal(goalId, payload);
    }

    return apiRequest<ApiResponse<DecomposeGoalResponse>>(
      `/api/goals/${goalId}/regenerate`,
      {
        method: "POST",
        body: JSON.stringify(payload)
      }
    ).then(unwrapData);
  },

  decomposeTask(taskId: string, payload: DecomposeGoalRequest) {
    if (USE_MOCKS) {
      return mockApi.decomposeTask(taskId, payload);
    }

    return apiRequest<ApiResponse<DecomposeGoalResponse>>(
      `/api/tasks/${taskId}/decompose`,
      {
        method: "POST",
        body: JSON.stringify(payload)
      }
    ).then(unwrapData);
  }
};

export const analyticsApi = {
  getProgress(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.getProgress(goalId);
    }

    return apiRequest<ApiResponse<ProgressResponse>>(
      `/api/goals/${goalId}/progress`
    ).then(unwrapData);
  },

  getFeasibility(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.getFeasibility(goalId);
    }

    return apiRequest<ApiResponse<FeasibilityResponse>>(
      `/api/goals/${goalId}/feasibility`
    ).then(unwrapData);
  },

  getQuality(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.getQuality(goalId);
    }

    return apiRequest<ApiResponse<QualityResponse>>(
      `/api/goals/${goalId}/quality`
    ).then(unwrapData);
  },

  getMetrics(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.getMetrics(goalId);
    }

    return apiRequest<ApiResponse<MetricsResponse>>(
      `/api/goals/${goalId}/metrics`
    ).then(unwrapData);
  }
};

export const planVersionsApi = {
  list(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.listPlanVersions(goalId);
    }

    return apiRequest<ApiResponse<PlanVersion[]>>(
      `/api/goals/${goalId}/versions`
    ).then(unwrapData);
  },

  get(goalId: string, versionId: string) {
    if (USE_MOCKS) {
      return mockApi.getPlanVersion(goalId, versionId);
    }

    return apiRequest<ApiResponse<PlanVersion>>(
      `/api/goals/${goalId}/versions/${versionId}`
    ).then(unwrapData);
  },

  activate(goalId: string, versionId: string) {
    if (USE_MOCKS) {
      return mockApi.activatePlanVersion(goalId, versionId);
    }

    return apiRequest<ApiResponse<PlanVersion>>(
      `/api/goals/${goalId}/versions/${versionId}/activate`,
      {
        method: "POST"
      }
    ).then(unwrapData);
  }
};

export const generationsApi = {
  listByGoal(goalId: string) {
    if (USE_MOCKS) {
      return mockApi.listGenerations(goalId);
    }

    return apiRequest<ApiResponse<Generation[]>>(
      `/api/goals/${goalId}/generations`
    ).then(unwrapData);
  },

  get(generationId: string) {
    if (USE_MOCKS) {
      return mockApi.getGeneration(generationId);
    }

    return apiRequest<ApiResponse<GenerationDetails>>(
      `/api/generations/${generationId}`
    ).then(unwrapData);
  }
};

export const feedbackApi = {
  create(generationId: string, payload: FeedbackRequest) {
    if (USE_MOCKS) {
      return mockApi.createFeedback(generationId, payload);
    }

    return apiRequest<ApiResponse<Feedback>>(
      `/api/generations/${generationId}/feedback`,
      {
        method: "POST",
        body: JSON.stringify(payload)
      }
    ).then(unwrapData);
  }
};

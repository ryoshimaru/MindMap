package http

import (
	"net/http"

	"github.com/ryoshimaru/MindMap/internal/http/handlers"
	"github.com/ryoshimaru/MindMap/internal/http/middleware"
)

func NewRouter(handler *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handler.Health)
	mux.HandleFunc("POST /api/auth/register", handler.Register)
	mux.HandleFunc("POST /api/auth/login", handler.Login)
	mux.HandleFunc("GET /api/auth/me", handler.CurrentUser)
	mux.HandleFunc("GET /api/profile", handler.GetProfile)
	mux.HandleFunc("PUT /api/profile", handler.SaveProfile)
	mux.HandleFunc("GET /api/auth/oauth/google/start", handler.StartGoogleOAuth)
	mux.HandleFunc("GET /api/auth/oauth/google/callback", handler.GoogleOAuthCallback)
	mux.HandleFunc("GET /api/ai-settings", handler.GetAISettings)
	mux.HandleFunc("PUT /api/ai-settings", handler.SaveAISettings)

	mux.HandleFunc("POST /api/requests/analyze", handler.AnalyzeRequest)
	mux.HandleFunc("POST /api/requests/{requestId}/answers", handler.SubmitRequestAnswers)
	mux.HandleFunc("POST /api/requests/{requestId}/evaluate", handler.EvaluateRequest)
	mux.HandleFunc("POST /api/requests/{requestId}/generate-plan", handler.GeneratePlanFromRequest)

	mux.HandleFunc("GET /api/goals", handler.ListGoals)
	mux.HandleFunc("POST /api/goals", handler.CreateGoal)
	mux.HandleFunc("GET /api/goals/{goalId}", handler.GetGoal)
	mux.HandleFunc("PUT /api/goals/{goalId}", handler.UpdateGoal)
	mux.HandleFunc("DELETE /api/goals/{goalId}", handler.DeleteGoal)
	mux.HandleFunc("GET /api/goals/{goalId}/context", handler.GetGoalContext)
	mux.HandleFunc("POST /api/goals/{goalId}/context", handler.CreateGoalContext)
	mux.HandleFunc("PUT /api/goals/{goalId}/context", handler.UpdateGoalContext)

	mux.HandleFunc("GET /api/goals/{goalId}/tasks", handler.ListGoalTasks)
	mux.HandleFunc("POST /api/goals/{goalId}/tasks", handler.CreateTask)
	mux.HandleFunc("GET /api/tasks/{taskId}", handler.GetTask)
	mux.HandleFunc("PUT /api/tasks/{taskId}", handler.UpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{taskId}", handler.DeleteTask)
	mux.HandleFunc("PATCH /api/tasks/{taskId}/status", handler.UpdateTaskStatus)
	mux.HandleFunc("GET /api/tasks/{taskId}/dependencies", handler.ListTaskDependencies)
	mux.HandleFunc("POST /api/tasks/{taskId}/dependencies", handler.CreateTaskDependency)
	mux.HandleFunc("DELETE /api/tasks/{taskId}/dependencies/{dependencyId}", handler.DeleteTaskDependency)

	mux.HandleFunc("POST /api/goals/{goalId}/decompose", handler.DecomposeGoal)
	mux.HandleFunc("POST /api/goals/{goalId}/regenerate", handler.RegenerateGoal)
	mux.HandleFunc("POST /api/tasks/{taskId}/decompose", handler.DecomposeTask)

	mux.HandleFunc("GET /api/goals/{goalId}/progress", handler.GetProgress)
	mux.HandleFunc("GET /api/goals/{goalId}/feasibility", handler.GetFeasibility)
	mux.HandleFunc("GET /api/goals/{goalId}/quality", handler.GetQuality)
	mux.HandleFunc("GET /api/goals/{goalId}/metrics", handler.GetMetrics)

	mux.HandleFunc("GET /api/goals/{goalId}/versions", handler.ListVersions)
	mux.HandleFunc("GET /api/goals/{goalId}/versions/{versionId}", handler.GetVersion)
	mux.HandleFunc("POST /api/goals/{goalId}/versions/{versionId}/activate", handler.ActivateVersion)
	mux.HandleFunc("GET /api/goals/{goalId}/generations", handler.ListGenerations)
	mux.HandleFunc("GET /api/generations/{generationId}", handler.GetGeneration)
	mux.HandleFunc("POST /api/generations/{generationId}/feedback", handler.CreateFeedback)

	mux.HandleFunc("/", handler.NotFound)

	return middleware.Recovery(middleware.Logging(middleware.CORS(mux)))
}

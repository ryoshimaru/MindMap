package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ryoshimaru/MindMap/internal/domain"
	"github.com/ryoshimaru/MindMap/internal/service"
	"github.com/ryoshimaru/MindMap/internal/store"
)

type Services struct {
	Auth      *service.AuthService
	Goals     *service.GoalService
	Tasks     *service.TaskService
	AI        *service.AIService
	Analytics *service.AnalyticsService
	Versions  *service.VersionService
}

type Handler struct {
	services Services
}

func New(services Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Сервер готов",
	})
}

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		WriteError(w, http.StatusNotFound, "not_found", "Адрес API не найден")
		return
	}
	http.NotFound(w, r)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body domain.RegisterRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	response, err := h.services.Auth.Register(body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body domain.LoginRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	response, err := h.services.Auth.Login(body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) CurrentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	profile, err := h.services.Auth.GetProfile(user.ID)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) SaveProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.UpsertUserProfileRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	profile, err := h.services.Auth.SaveProfile(user.ID, body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) StartGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	if clientID == "" || clientSecret == "" {
		WriteError(w, http.StatusBadRequest, "oauth_not_configured", "Вход через Google не настроен")
		return
	}

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("redirect_uri", googleRedirectURL(r))
	values.Set("response_type", "code")
	values.Set("scope", "openid email profile")
	values.Set("access_type", "offline")
	values.Set("prompt", "select_account")
	http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+values.Encode(), http.StatusFound)
}

func (h *Handler) GoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")) == "" || strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")) == "" {
		WriteError(w, http.StatusBadRequest, "oauth_not_configured", "Вход через Google не настроен")
		return
	}

	token, err := h.exchangeGoogleCode(r)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	http.Redirect(w, r, frontendOAuthCallbackURL(token), http.StatusFound)
}

func (h *Handler) exchangeGoogleCode(r *http.Request) (string, error) {
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		return "", store.ErrValidation
	}

	values := url.Values{}
	values.Set("code", code)
	values.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	values.Set("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	values.Set("redirect_uri", googleRedirectURL(r))
	values.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", values)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 || tokenResponse.AccessToken == "" {
		return "", store.ErrValidation
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	userResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer userResp.Body.Close()

	var googleUser struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&googleUser); err != nil {
		return "", err
	}
	if userResp.StatusCode >= 400 || strings.TrimSpace(googleUser.Email) == "" {
		return "", store.ErrValidation
	}

	return h.services.Auth.CreateOAuthUserSession(googleUser.Email, googleUser.Name, googleUser.Picture, domain.AuthProviderGoogle)
}

func googleRedirectURL(r *http.Request) string {
	if value := strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL")); value != "" {
		return value
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/api/auth/oauth/google/callback"
}

func frontendOAuthCallbackURL(token string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	values := url.Values{}
	values.Set("token", token)
	return baseURL + "/auth/callback?" + values.Encode()
}

func (h *Handler) GetAISettings(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	settings, err := h.services.AI.GetSettings(user.ID)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, settings)
}

func (h *Handler) SaveAISettings(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.AISettingsRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	settings, err := h.services.AI.SaveSettings(user.ID, body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, settings)
}

func (h *Handler) AnalyzeRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.AnalyzeFreeformRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	response, err := h.services.AI.AnalyzeFreeform(user.ID, body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) SubmitRequestAnswers(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.SubmitRequestAnswers
	if !DecodeJSON(w, r, &body) {
		return
	}
	response, err := h.services.AI.SaveAnswers(user.ID, r.PathValue("requestId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) EvaluateRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	response, err := h.services.AI.EvaluateRequest(user.ID, r.PathValue("requestId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GeneratePlanFromRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	response, err := h.services.AI.GeneratePlan(user.ID, r.PathValue("requestId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) ListGoals(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	WriteJSON(w, http.StatusOK, h.services.Goals.List(user.ID))
}

func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.CreateGoalRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	goal, err := h.services.Goals.Create(user.ID, body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, goal)
}

func (h *Handler) GetGoal(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	goal, err := h.services.Goals.Get(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, goal)
}

func (h *Handler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.UpdateGoalRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	goal, err := h.services.Goals.Update(user.ID, r.PathValue("goalId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, goal)
}

func (h *Handler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	if err := h.services.Goals.Delete(user.ID, r.PathValue("goalId")); err != nil {
		WriteStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetGoalContext(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	context, err := h.services.Goals.GetContext(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, context)
}

func (h *Handler) CreateGoalContext(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.UpsertGoalContextRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	context, err := h.services.Goals.CreateContext(user.ID, r.PathValue("goalId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, context)
}

func (h *Handler) UpdateGoalContext(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.UpsertGoalContextRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	context, err := h.services.Goals.UpdateContext(user.ID, r.PathValue("goalId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, context)
}

func (h *Handler) ListGoalTasks(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	tasks, err := h.services.Tasks.ListByGoal(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, tasks)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.CreateTaskRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	task, err := h.services.Tasks.Create(user.ID, r.PathValue("goalId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, task)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	task, err := h.services.Tasks.Get(user.ID, r.PathValue("taskId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, task)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.UpdateTaskRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	task, err := h.services.Tasks.Update(user.ID, r.PathValue("taskId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, task)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	if err := h.services.Tasks.Delete(user.ID, r.PathValue("taskId")); err != nil {
		WriteStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.UpdateTaskStatusRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	task, err := h.services.Tasks.UpdateStatus(user.ID, r.PathValue("taskId"), body.Status)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, task)
}

func (h *Handler) ListTaskDependencies(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	deps, err := h.services.Tasks.ListDependencies(user.ID, r.PathValue("taskId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, deps)
}

func (h *Handler) CreateTaskDependency(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.CreateTaskDependencyRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	dep, err := h.services.Tasks.CreateDependency(user.ID, r.PathValue("taskId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, dep)
}

func (h *Handler) DeleteTaskDependency(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	if err := h.services.Tasks.DeleteDependency(user.ID, r.PathValue("taskId"), r.PathValue("dependencyId")); err != nil {
		WriteStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DecomposeGoal(w http.ResponseWriter, r *http.Request) {
	h.decomposeGoal(w, r, "initial")
}

func (h *Handler) RegenerateGoal(w http.ResponseWriter, r *http.Request) {
	h.decomposeGoal(w, r, "regenerate")
}

func (h *Handler) decomposeGoal(w http.ResponseWriter, r *http.Request, mode string) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.DecomposeGoalRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	response, err := h.services.AI.DecomposeGoal(user.ID, r.PathValue("goalId"), body, mode)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) DecomposeTask(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.DecomposeGoalRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	response, err := h.services.AI.DecomposeTask(user.ID, r.PathValue("taskId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	response, err := h.services.Analytics.Progress(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetFeasibility(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	response, err := h.services.Analytics.Feasibility(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetQuality(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	response, err := h.services.Analytics.Quality(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	response, err := h.services.Analytics.Metrics(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	versions, err := h.services.Versions.List(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, versions)
}

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	version, err := h.services.Versions.Get(user.ID, r.PathValue("goalId"), r.PathValue("versionId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, version)
}

func (h *Handler) ActivateVersion(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	version, err := h.services.Versions.Activate(user.ID, r.PathValue("goalId"), r.PathValue("versionId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, version)
}

func (h *Handler) ListGenerations(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	generations, err := h.services.AI.ListGenerations(user.ID, r.PathValue("goalId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, generations)
}

func (h *Handler) GetGeneration(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	generation, err := h.services.AI.GetGeneration(user.ID, r.PathValue("generationId"))
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, generation)
}

func (h *Handler) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	user, ok := h.RequireUser(w, r)
	if !ok {
		return
	}
	var body domain.FeedbackRequest
	if !DecodeJSON(w, r, &body) {
		return
	}
	feedback, err := h.services.AI.CreateFeedback(user.ID, r.PathValue("generationId"), body)
	if err != nil {
		WriteStoreError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, feedback)
}

func (h *Handler) RequireUser(w http.ResponseWriter, r *http.Request) (domain.User, bool) {
	token := BearerToken(r)
	if token == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
		return domain.User{}, false
	}
	user, err := h.services.Auth.UserByToken(token)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Сессия истекла или недействительна")
		return domain.User{}, false
	}
	return user, true
}

func BearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return ""
	}
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok {
		return ""
	}
	return strings.TrimSpace(token)
}

package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ryoshimaru/MindMap/internal/domain"
)

type Provider interface {
	Name() string
	AnalyzeRequest(text string, now time.Time) (domain.AnalyzeFreeformResponse, error)
	EvaluateRequest(request domain.PlanningRequest) (domain.EvaluateRequestResponse, error)
	GeneratePlan(request domain.PlanningRequest) (domain.GeneratedPlan, error)
}

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) AnalyzeRequest(text string, now time.Time) (domain.AnalyzeFreeformResponse, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return domain.AnalyzeFreeformResponse{}, fmt.Errorf("empty request")
	}

	return domain.AnalyzeFreeformResponse{
		OriginalText: trimmed,
		InterpretedGoal: domain.InterpretedGoal{
			Title:       firstLine(trimmed),
			Description: trimmed,
			Category:    "General",
			Deadline:    nil,
			Constraints: []string{},
		},
		Feasibility: domain.RequestFeasibility{
			Status:                 domain.RequestFeasibilityNeedsClarification,
			Reason:                 "The mock provider requires generic clarification before roadmap generation.",
			BlockingFactors:        []string{"Missing structured context."},
			Assumptions:            []string{},
			RequiredClarifications: []string{"current_state", "available_time", "success_criteria"},
			CanGeneratePlan:        false,
		},
		ClarifyingQuestions: genericMockQuestions(),
		ActivityTracker: domain.ActivityTracker{
			Stage:      "analysis_ready",
			Confidence: 70,
			Items: []domain.ActivitySignal{
				{Label: "Provider", Value: "mock"},
				{Label: "Mode", Value: "universal structured analysis"},
			},
		},
	}, nil
}

func (p *MockProvider) EvaluateRequest(request domain.PlanningRequest) (domain.EvaluateRequestResponse, error) {
	missing := requiredMissing(request.Answers, "current_state", "available_time", "success_criteria")
	feasibility := domain.RequestFeasibility{
		Status:                 domain.RequestFeasibilityFeasible,
		Reason:                 "The mock provider has enough generic context to generate a demo roadmap.",
		BlockingFactors:        []string{},
		Assumptions:            []string{"Mock evaluation does not perform domain-specific judgment."},
		RequiredClarifications: []string{},
		CanGeneratePlan:        true,
	}
	questions := []domain.ClarifyingQuestion{}
	if len(missing) > 0 {
		feasibility = domain.RequestFeasibility{
			Status:                 domain.RequestFeasibilityNeedsClarification,
			Reason:                 "Some required generic clarification answers are still missing.",
			BlockingFactors:        missing,
			Assumptions:            []string{},
			RequiredClarifications: missing,
			CanGeneratePlan:        false,
		}
		questions = questionsByID(genericMockQuestions(), missing)
	}

	return domain.EvaluateRequestResponse{
		RequestID:           request.ID,
		Feasibility:         feasibility,
		ClarifyingQuestions: questions,
		ActivityTracker: domain.ActivityTracker{
			Stage:      "feasibility_evaluated",
			Confidence: 75,
			Items: []domain.ActivitySignal{
				{Label: "Can generate roadmap", Value: fmt.Sprintf("%t", feasibility.CanGeneratePlan)},
			},
		},
	}, nil
}

func (p *MockProvider) GeneratePlan(request domain.PlanningRequest) (domain.GeneratedPlan, error) {
	result := answerString(request.Answers, "success_criteria", "A concrete first outcome for the requested goal.")
	currentState := answerString(request.Answers, "current_state", "Current state was not specified.")
	availableTime := answerString(request.Answers, "available_time", "Regular focused work sessions.")

	return domain.GeneratedPlan{
		Goal: domain.CreateGoalRequest{
			Title:                 request.InterpretedGoal.Title,
			Description:           request.InterpretedGoal.Description,
			Category:              request.InterpretedGoal.Category,
			Priority:              domain.GoalPriorityHigh,
			Deadline:              request.InterpretedGoal.Deadline,
			AvailableHoursPerWeek: 10,
		},
		Context: domain.UpsertGoalContextRequest{
			CurrentLevel:    currentState,
			ConstraintsText: answerString(request.Answers, "constraints", ""),
			Preferences:     []string{availableTime},
			ExpectedResult:  result,
		},
		Tasks: []domain.GeneratedPlanTask{
			{
				Title:          "Clarify measurable outcome",
				Description:    "Define the success criteria, scope, and first milestone.",
				Priority:       domain.GoalPriorityHigh,
				EstimatedHours: 1,
				Deadline:       request.InterpretedGoal.Deadline,
			},
			{
				Title:          "Prepare resources and constraints",
				Description:    "Collect required materials, schedule available time, and resolve blockers.",
				Priority:       domain.GoalPriorityMedium,
				EstimatedHours: 2,
				Deadline:       request.InterpretedGoal.Deadline,
			},
			{
				Title:          "Execute first roadmap milestone",
				Description:    result,
				Priority:       domain.GoalPriorityHigh,
				EstimatedHours: 4,
				Deadline:       request.InterpretedGoal.Deadline,
			},
		},
	}, nil
}

func genericMockQuestions() []domain.ClarifyingQuestion {
	minZero := 0.0
	return []domain.ClarifyingQuestion{
		{ID: "current_state", Text: "Опишите текущую стартовую точку для этой цели.", Type: domain.QuestionTypeTextarea, Required: true, Placeholder: stringPointer("Ресурсы, навыки, ограничения или контекст")},
		{ID: "available_time", Text: "Сколько времени или усилий вы реально можете выделять?", Type: domain.QuestionTypeText, Required: true, Placeholder: stringPointer("Например: 5 часов в неделю")},
		{ID: "success_criteria", Text: "Какой результат будет считаться успехом?", Type: domain.QuestionTypeTextarea, Required: true},
		{ID: "target_date", Text: "Есть ли целевая дата?", Type: domain.QuestionTypeDate, Required: false},
		{ID: "risk_tolerance", Text: "Какой уровень риска допустим?", Type: domain.QuestionTypeSingleSelect, Required: false, Options: []domain.ClarifyingQuestionOption{{Value: "low", Label: "Низкий"}, {Value: "medium", Label: "Средний"}, {Value: "high", Label: "Высокий"}}},
		{ID: "important_constraints", Text: "Какие типы ограничений важнее всего?", Type: domain.QuestionTypeMultiSelect, Required: false, Options: []domain.ClarifyingQuestionOption{{Value: "time", Label: "Время"}, {Value: "money", Label: "Деньги"}, {Value: "risk", Label: "Риски"}, {Value: "access", Label: "Доступ"}}},
		{ID: "has_required_resources", Text: "У вас уже есть основные необходимые ресурсы?", Type: domain.QuestionTypeBoolean, Required: false},
		{ID: "budget", Text: "Если актуально, какой бюджет доступен?", Type: domain.QuestionTypeNumber, Required: false, Min: &minZero, Unit: stringPointer("валюта")},
	}
}

func requiredMissing(answers []domain.RequestAnswer, questionIDs ...string) []string {
	answerByID := make(map[string]domain.RequestAnswer, len(answers))
	for _, answer := range answers {
		answerByID[answer.QuestionID] = answer
	}
	missing := make([]string, 0)
	for _, questionID := range questionIDs {
		answer, ok := answerByID[questionID]
		if !ok || len(answer.Value) == 0 || string(answer.Value) == `""` {
			missing = append(missing, questionID)
		}
	}
	return missing
}

func questionsByID(questions []domain.ClarifyingQuestion, ids []string) []domain.ClarifyingQuestion {
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	result := make([]domain.ClarifyingQuestion, 0, len(ids))
	for _, question := range questions {
		if want[question.ID] {
			result = append(result, question)
		}
	}
	return result
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	line = strings.TrimSpace(line)
	if line == "" {
		return "New goal"
	}
	return line
}

func answerString(answers []domain.RequestAnswer, questionID, fallback string) string {
	for _, answer := range answers {
		if answer.QuestionID != questionID {
			continue
		}
		var value string
		if err := json.Unmarshal(answer.Value, &value); err == nil && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return fallback
}

func stringPointer(value string) *string {
	return &value
}

package ai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryoshimaru/MindMap/internal/domain"
)

func providerAnalyzePrompt(text string, now time.Time) string {
	return fmt.Sprintf(`Analyze the user's planning request and return strict JSON.

Current date: %s
User request: %q

Return this JSON shape:
{
  "interpretedGoal": {
    "title": "short actionable goal title",
    "description": "expanded goal description",
    "category": "category name",
    "deadline": "YYYY-MM-DD or null",
    "constraints": ["constraint 1"]
  },
  "feasibility": {
    "status": "UNKNOWN|NEEDS_CLARIFICATION|FEASIBLE|RISKY|INFEASIBLE|UNSAFE",
    "reason": "short explanation",
    "blockingFactors": ["factor"],
    "assumptions": ["assumption"],
    "requiredClarifications": ["question_id"],
    "canGeneratePlan": false
  },
  "clarifyingQuestions": [
    {
      "id": "stable_snake_case_id",
      "text": "question for UI",
      "type": "text|textarea|number|date|single_select|multi_select|boolean",
      "required": true,
      "options": [{"value":"stable_value","label":"Human label"}],
      "min": 0,
      "max": 100,
      "unit": "hours",
      "placeholder": "optional input hint"
    }
  ],
  "activityTracker": {
    "stage": "analysis_ready",
    "confidence": 0,
    "items": [{"label":"Detected category","value":"..."}]
  }
}

Rules:
- This is not a chat. clarifyingQuestions must be structured controls for UI.
- Always evaluate whether enough data exists to judge feasibility before planning.
- If data is missing, set feasibility.status to NEEDS_CLARIFICATION and ask targeted questions.
- Set feasibility.canGeneratePlan to true only when roadmap generation is allowed; otherwise false.
- Consider physical, financial, time, legal, and safety feasibility.
- For risky purchases, health/body goals, illegal, dangerous, or physically impossible requests, do not pretend they are feasible.
- Include question types from: text, textarea, number, date, single_select, multi_select, boolean.
- Put IDs of still-needed answers in feasibility.requiredClarifications.
- Use null for unknown deadline.
- Use Russian for user-facing question text when the user request is Russian.
- Do not invent a requestId.`, now.Format("2006-01-02"), text)
}

func providerEvaluatePrompt(request domain.PlanningRequest) string {
	payload, _ := json.MarshalIndent(request, "", "  ")
	return fmt.Sprintf(`Re-evaluate feasibility after the user's structured answers.

Planning request JSON:
%s

Return strict JSON:
{
  "feasibility": {
    "status": "UNKNOWN|NEEDS_CLARIFICATION|FEASIBLE|RISKY|INFEASIBLE|UNSAFE",
    "reason": "clear explanation",
    "blockingFactors": ["factor"],
    "assumptions": ["assumption"],
    "requiredClarifications": ["question_id"],
    "canGeneratePlan": false
  },
  "clarifyingQuestions": [
    {
      "id": "stable_snake_case_id",
      "text": "question for UI",
      "type": "text|textarea|number|date|single_select|multi_select|boolean",
      "required": true,
      "options": [{"value":"stable_value","label":"Human label"}],
      "min": 0,
      "max": 100,
      "unit": "unit",
      "placeholder": "optional input hint"
    }
  ],
  "activityTracker": {
    "stage": "feasibility_evaluated",
    "confidence": 0,
    "items": [{"label":"Decision","value":"..."}]
  }
}

Rules:
- If the request is impossible, unsafe, illegal, or financially/physically impossible, set INFEASIBLE or UNSAFE.
- If key data is still missing, set NEEDS_CLARIFICATION and return only the missing questions.
- Put IDs of still-needed answers in feasibility.requiredClarifications.
- Set canGeneratePlan to true only when a task plan may reasonably be generated. FEASIBLE and RISKY may allow generation; all other statuses must set canGeneratePlan to false.
- Do not include prose outside JSON.`, string(payload))
}

func providerPlanPrompt(request domain.PlanningRequest) string {
	payload, _ := json.MarshalIndent(request, "", "  ")
	return fmt.Sprintf(`Generate a practical plan from this structured planning request.

Planning request JSON:
%s

Return this strict JSON shape:
{
  "goal": {
    "title": "goal title",
    "description": "goal description",
    "category": "category",
    "priority": "LOW|MEDIUM|HIGH",
    "deadline": "YYYY-MM-DD or null",
    "availableHoursPerWeek": 10
  },
  "context": {
    "currentLevel": "current user level",
    "constraintsText": "constraints summary",
    "preferences": ["preference"],
    "expectedResult": "concrete expected result"
  },
  "tasks": [
    {
      "title": "task title",
      "description": "task description",
      "priority": "LOW|MEDIUM|HIGH",
      "estimatedHours": 2,
      "deadline": "YYYY-MM-DD or null",
      "children": []
    }
  ]
}

Rules:
- Produce a real task tree, not prose.
- Only generate a plan if feasibility.canGeneratePlan is true. Otherwise return an error JSON object is not allowed; the backend should block before this prompt.
- Make tasks concrete, sequenced, and feasible for the deadline.
- Use children for meaningful subtasks.
- Keep priority enum values exactly LOW, MEDIUM, HIGH.
- Use null for unknown deadlines.`, string(payload))
}

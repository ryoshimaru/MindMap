import { useMemo, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { requestsApi } from "../api/endpoints";
import type {
  ActivityTracker,
  AnalyzeGoalResponse,
  ClarifyingQuestion,
  EvaluateGoalRequestResponse,
  InterpretedGoal,
  RequestAnswerValue,
  RequestFeasibility
} from "../api/types";
import { getErrorMessage } from "../utils/errors";
import { ErrorState } from "./ErrorState";

type SmartGoalStep = "idle" | "analyzing" | "answering" | "evaluating" | "generating";
type AnswersState = Record<string, RequestAnswerValue>;

interface RequestWorkspaceState {
  requestId: string;
  interpretedGoal: InterpretedGoal;
  feasibility: RequestFeasibility;
  clarifyingQuestions: ClarifyingQuestion[];
  activityTracker: ActivityTracker;
}

const stageLabels: Record<string, string> = {
  analysis_ready: "Анализ готов",
  questions_ready: "Вопросы готовы",
  answers_saved: "Ответы сохранены",
  feasibility_evaluated: "Оценка готова",
  plan_generated: "Roadmap построен"
};

const feasibilityLabels: Record<string, string> = {
  UNKNOWN: "Недостаточно данных",
  NEEDS_CLARIFICATION: "Нужны уточнения",
  FEASIBLE: "Можно строить roadmap",
  RISKY: "Есть риски",
  INFEASIBLE: "Цель сейчас нереализуема",
  UNSAFE: "Запрос небезопасен"
};

function normalizeQuestionValue(question: ClarifyingQuestion): RequestAnswerValue {
  if (question.type === "multi_select") return [];
  if (question.type === "boolean") return false;
  if (question.type === "number") return 0;
  return "";
}

function isAnswered(value: RequestAnswerValue) {
  if (Array.isArray(value)) return value.length > 0;
  if (typeof value === "boolean") return true;
  return String(value).trim() !== "";
}

function mergeAnswerDefaults(questions: ClarifyingQuestion[], currentAnswers: AnswersState) {
  return Object.fromEntries(
    questions.map((question) => [
      question.id,
      currentAnswers[question.id] ?? normalizeQuestionValue(question)
    ])
  );
}

function QuestionField({
  question,
  value,
  onChange
}: {
  question: ClarifyingQuestion;
  value: RequestAnswerValue;
  onChange: (value: RequestAnswerValue) => void;
}) {
  const label = question.text;

  if (question.type === "textarea") {
    return (
      <textarea
        className="field-control"
        value={String(value)}
        required={question.required}
        placeholder={question.placeholder ?? undefined}
        aria-label={label}
        onChange={(event) => onChange(event.target.value)}
      />
    );
  }

  if (question.type === "single_select") {
    return (
      <select
        className="field-control"
        value={String(value)}
        required={question.required}
        aria-label={label}
        onChange={(event) => onChange(event.target.value)}
      >
        <option value="">Выберите вариант</option>
        {(question.options || []).map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    );
  }

  if (question.type === "multi_select") {
    const selectedValues = Array.isArray(value) ? value : [];
    return (
      <div className="choice-list" role="group" aria-label={label}>
        {(question.options || []).map((option) => (
          <label key={option.value} className="choice-list__item">
            <input
              type="checkbox"
              checked={selectedValues.includes(option.value)}
              onChange={(event) =>
                onChange(
                  event.target.checked
                    ? [...selectedValues, option.value]
                    : selectedValues.filter((item) => item !== option.value)
                )
              }
            />
            <span>{option.label}</span>
          </label>
        ))}
      </div>
    );
  }

  if (question.type === "boolean") {
    return (
      <label className="choice-list__item">
        <input
          type="checkbox"
          checked={Boolean(value)}
          onChange={(event) => onChange(event.target.checked)}
        />
        <span>{question.placeholder || "Да"}</span>
      </label>
    );
  }

  return (
    <input
      className="field-control"
      type={question.type === "number" ? "number" : question.type}
      value={String(value)}
      min={question.min ?? undefined}
      max={question.max ?? undefined}
      required={question.required}
      placeholder={question.placeholder ?? undefined}
      aria-label={label}
      onChange={(event) =>
        onChange(question.type === "number" ? Number(event.target.value) : event.target.value)
      }
    />
  );
}

function FeasibilityPanel({ feasibility }: { feasibility: RequestFeasibility }) {
  const statusLabel = feasibilityLabels[feasibility.status] || feasibility.status;
  const isBlocking = feasibility.canGeneratePlan === false;

  return (
    <section className={`metric-card feasibility-panel${isBlocking ? " feasibility-panel--blocked" : ""}`}>
      <div className="card-header">
        <span className="eyebrow">Оценка реализуемости</span>
        <strong>{statusLabel}</strong>
      </div>
      {feasibility.reason ? <p>{feasibility.reason}</p> : null}
      {feasibility.blockingFactors.length > 0 ? (
        <ul className="api-checklist">
          {feasibility.blockingFactors.map((factor) => (
            <li key={factor}>{factor}</li>
          ))}
        </ul>
      ) : null}
      {feasibility.status === "RISKY" && feasibility.canGeneratePlan ? (
        <p className="field-hint">AI разрешил построить roadmap, но отметил риски.</p>
      ) : null}
    </section>
  );
}

function ActivityPanel({ tracker }: { tracker: ActivityTracker }) {
  return (
    <aside className="metric-card">
      <span className="eyebrow">Ход работы</span>
      <div className="activity-tracker__summary">
        <strong>{stageLabels[tracker.stage] || tracker.stage}</strong>
        <span>{tracker.confidence}%</span>
      </div>
      <ol className="activity-tracker" aria-label="Статус анализа цели">
        {tracker.items.map((item) => (
          <li key={`${item.label}-${item.value}`} className="activity-tracker__item">
            <span className="activity-tracker__marker" aria-hidden="true" />
            <div>
              <strong>{item.label}</strong>
              <p>{item.value}</p>
            </div>
          </li>
        ))}
      </ol>
    </aside>
  );
}

function normalizeFeasibility(feasibility: RequestFeasibility): RequestFeasibility {
  return {
    status: feasibility.status || "UNKNOWN",
    reason: feasibility.reason || "",
    blockingFactors: feasibility.blockingFactors || [],
    assumptions: feasibility.assumptions || [],
    requiredClarifications: feasibility.requiredClarifications || [],
    canGeneratePlan: feasibility.canGeneratePlan === true
  };
}

function createWorkspaceState(response: AnalyzeGoalResponse): RequestWorkspaceState {
  return {
    requestId: response.requestId,
    interpretedGoal: response.interpretedGoal,
    feasibility: normalizeFeasibility(response.feasibility),
    clarifyingQuestions: response.clarifyingQuestions || [],
    activityTracker: response.activityTracker
  };
}

function applyEvaluation(current: RequestWorkspaceState, response: EvaluateGoalRequestResponse): RequestWorkspaceState {
  return {
    ...current,
    feasibility: normalizeFeasibility(response.feasibility),
    clarifyingQuestions: response.clarifyingQuestions || [],
    activityTracker: response.activityTracker
  };
}

export function SmartGoalRequestPanel() {
  const navigate = useNavigate();
  const [text, setText] = useState("");
  const [workspace, setWorkspace] = useState<RequestWorkspaceState | null>(null);
  const [answers, setAnswers] = useState<AnswersState>({});
  const [step, setStep] = useState<SmartGoalStep>("idle");
  const [errorMessage, setErrorMessage] = useState("");

  const requiredQuestions = useMemo(
    () => workspace?.clarifyingQuestions.filter((question) => question.required) || [],
    [workspace]
  );
  const canGeneratePlan = workspace?.feasibility.canGeneratePlan === true;
  const hasQuestions = Boolean(workspace?.clarifyingQuestions.length);
  const canSubmitAnswers = hasQuestions && step !== "evaluating" && step !== "generating";
  const canShowRoadmapButton = Boolean(workspace && canGeneratePlan);

  async function handleAnalyze(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStep("analyzing");
    setErrorMessage("");
    setWorkspace(null);
    setAnswers({});

    try {
      const result = await requestsApi.analyze({ text });
      const nextWorkspace = createWorkspaceState(result);
      setWorkspace(nextWorkspace);
      setAnswers(mergeAnswerDefaults(nextWorkspace.clarifyingQuestions, {}));
      setStep("answering");
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
      setStep("idle");
    }
  }

  async function handleSubmitAnswers(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!workspace) return;

    const missingQuestion = requiredQuestions.find((question) => !isAnswered(answers[question.id]));
    if (missingQuestion) {
      setErrorMessage(`Ответьте на обязательный вопрос: ${missingQuestion.text}`);
      return;
    }

    setStep("evaluating");
    setErrorMessage("");

    try {
      await requestsApi.submitAnswers(workspace.requestId, {
        answers: workspace.clarifyingQuestions.map((question) => ({
          questionId: question.id,
          value: answers[question.id]
        }))
      });
      const evaluation = await requestsApi.evaluate(workspace.requestId);
      const nextWorkspace = applyEvaluation(workspace, evaluation);
      setWorkspace(nextWorkspace);
      setAnswers((currentAnswers) => mergeAnswerDefaults(nextWorkspace.clarifyingQuestions, currentAnswers));
      setStep("answering");
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
      setStep("answering");
    }
  }

  async function handleGeneratePlan() {
    if (!workspace || workspace.feasibility.canGeneratePlan !== true) return;

    setStep("generating");
    setErrorMessage("");

    try {
      const generatedPlan = await requestsApi.generatePlan(workspace.requestId);
      navigate(`/goals/${generatedPlan.goal.id}`);
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
      setStep("answering");
    }
  }

  const isBusy = step === "analyzing" || step === "evaluating" || step === "generating";

  return (
    <section className="card card--primary-panel smart-request-panel">
      <div className="page-section__header">
        <div>
          <span className="eyebrow">Новая цель</span>
          <h2>Что вы хотите сделать?</h2>
        </div>
      </div>

      <form className="form-grid" onSubmit={handleAnalyze}>
        <div className="field-group">
          <label className="field-label" htmlFor="goal-request-text">
            Свободная цель
          </label>
          <textarea
            id="goal-request-text"
            className="field-control smart-request-panel__prompt"
            value={text}
            placeholder="Опишите цель любым текстом"
            required
            onChange={(event) => setText(event.target.value)}
          />
        </div>

        <div className="inline-form-actions">
          <button type="submit" className="button button--primary" disabled={isBusy}>
            {step === "analyzing" ? "Анализ..." : "Проанализировать"}
          </button>
          <span className="field-hint">Вопросы, оценку и разрешение на roadmap возвращает backend/AI.</span>
        </div>
      </form>

      {errorMessage ? <ErrorState description={errorMessage} /> : null}

      {workspace ? (
        <div className="smart-request-panel__workspace">
          <section className="interpretation-box">
            <span className="eyebrow">Интерпретация AI</span>
            <h3>{workspace.interpretedGoal.title}</h3>
            <p>{workspace.interpretedGoal.description}</p>
            <dl className="interpretation-box__meta">
              <div>
                <dt>Категория</dt>
                <dd>{workspace.interpretedGoal.category}</dd>
              </div>
              <div>
                <dt>Срок</dt>
                <dd>{workspace.interpretedGoal.deadline || "не указан"}</dd>
              </div>
              <div>
                <dt>Разрешение roadmap</dt>
                <dd>{canGeneratePlan ? "разрешено" : "не разрешено"}</dd>
              </div>
            </dl>
          </section>

          <div className="smart-request-panel__grid">
            <div className="section-stack">
              <FeasibilityPanel feasibility={workspace.feasibility} />

              {hasQuestions ? (
                <form className="form-grid" onSubmit={handleSubmitAnswers}>
                  <div>
                    <span className="eyebrow">Уточняющие вопросы</span>
                    <h3>Ответьте на поля от AI</h3>
                  </div>

                  {workspace.clarifyingQuestions.map((question) => (
                    <div className="field-group" key={question.id}>
                      <label className="field-label">
                        {question.text}
                        {question.unit ? `, ${question.unit}` : ""}
                        {question.required ? " *" : ""}
                      </label>
                      <QuestionField
                        question={question}
                        value={answers[question.id] ?? normalizeQuestionValue(question)}
                        onChange={(value) =>
                          setAnswers((currentAnswers) => ({
                            ...currentAnswers,
                            [question.id]: value
                          }))
                        }
                      />
                    </div>
                  ))}

                  <button type="submit" className="button button--secondary" disabled={!canSubmitAnswers}>
                    {step === "evaluating" ? "Оценка..." : "Отправить ответы"}
                  </button>
                </form>
              ) : null}

              {canShowRoadmapButton ? (
                <button
                  type="button"
                  className="button button--primary"
                  disabled={step === "generating"}
                  onClick={() => void handleGeneratePlan()}
                >
                  {step === "generating" ? "Построение..." : "Построить roadmap"}
                </button>
              ) : null}
            </div>

            <ActivityPanel tracker={workspace.activityTracker} />
          </div>
        </div>
      ) : null}
    </section>
  );
}

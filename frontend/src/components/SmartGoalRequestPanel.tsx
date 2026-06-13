import { useEffect, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { requestsApi } from "../api/endpoints";
import type { AnalyzeGoalResponse, ClarifyingQuestion, RequestAnswerValue, RequestFeasibility } from "../api/types";
import { getErrorMessage } from "../utils/errors";

const DRAFT_KEY = "mindmap.goal-draft";
type Answers = Record<string, RequestAnswerValue>;

function emptyValue(question: ClarifyingQuestion): RequestAnswerValue {
  if (question.type === "multi_select") return [];
  if (question.type === "boolean") return false;
  return "";
}

function QuestionInput({ question, value, onChange }: { question: ClarifyingQuestion; value: RequestAnswerValue; onChange: (value: RequestAnswerValue) => void }) {
  if (question.type === "textarea") return <textarea autoFocus value={String(value)} placeholder={question.placeholder || "Ваш ответ"} onChange={(e) => onChange(e.target.value)} />;
  if (question.type === "single_select") return <div className="answer-options">{question.options?.map((option) => <button type="button" className={value === option.value ? "selected" : ""} key={option.value} onClick={() => onChange(option.value)}>{option.label}</button>)}</div>;
  if (question.type === "multi_select") {
    const selected = Array.isArray(value) ? value : [];
    return <div className="answer-options">{question.options?.map((option) => <button type="button" className={selected.includes(option.value) ? "selected" : ""} key={option.value} onClick={() => onChange(selected.includes(option.value) ? selected.filter((v) => v !== option.value) : [...selected, option.value])}>{option.label}</button>)}</div>;
  }
  if (question.type === "boolean") return <div className="answer-options"><button type="button" className={value === true ? "selected" : ""} onClick={() => onChange(true)}>Да</button><button type="button" className={value === false ? "selected" : ""} onClick={() => onChange(false)}>Нет</button></div>;
  return <input autoFocus type={question.type === "number" ? "number" : question.type} min={question.min ?? undefined} max={question.max ?? undefined} value={String(value)} placeholder={question.placeholder || "Ваш ответ"} onChange={(e) => onChange(question.type === "number" ? Number(e.target.value) : e.target.value)} />;
}

function answered(value: RequestAnswerValue | undefined) {
  return Array.isArray(value) ? value.length > 0 : typeof value === "boolean" ? true : String(value ?? "").trim().length > 0;
}

export function SmartGoalRequestPanel() {
  const navigate = useNavigate();
  const [text, setText] = useState(() => localStorage.getItem(DRAFT_KEY) || "");
  const [analysis, setAnalysis] = useState<AnalyzeGoalResponse | null>(null);
  const [answers, setAnswers] = useState<Answers>({});
  const [questionIndex, setQuestionIndex] = useState(0);
  const [feasibility, setFeasibility] = useState<RequestFeasibility | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => { localStorage.setItem(DRAFT_KEY, text); }, [text]);
  const questions = analysis?.clarifyingQuestions || [];
  const question = questions[questionIndex];
  const progress = questions.length ? ((questionIndex + 1) / questions.length) * 100 : 0;
  const canContinue = question ? !question.required || answered(answers[question.id]) : false;

  async function analyze(goalText: string) {
    setBusy(true); setError(""); setFeasibility(null);
    try {
      const result = await requestsApi.analyze({text: goalText});
      setAnalysis(result); setAnswers(Object.fromEntries(result.clarifyingQuestions.map((item) => [item.id, emptyValue(item)]))); setQuestionIndex(0);
      if (!result.clarifyingQuestions.length) setFeasibility(result.feasibility);
    } catch (reason) { setError(getErrorMessage(reason)); }
    finally { setBusy(false); }
  }

  async function start(event: FormEvent) { event.preventDefault(); await analyze(text); }
  async function continueQuestions() {
    if (!analysis || !question) return;
    if (questionIndex < questions.length - 1) { setQuestionIndex((value) => value + 1); return; }
    setBusy(true); setError("");
    try {
      await requestsApi.submitAnswers(analysis.requestId, {answers: questions.map((item) => ({questionId: item.id, value: answers[item.id]}))});
      const result = await requestsApi.evaluate(analysis.requestId);
      setFeasibility(result.feasibility);
    } catch (reason) { setError(getErrorMessage(reason)); }
    finally { setBusy(false); }
  }

  async function generate() {
    if (!analysis) return; setBusy(true); setError("");
    try { const result = await requestsApi.generatePlan(analysis.requestId); localStorage.removeItem(DRAFT_KEY); navigate(`/goals/${result.goal.id}`); }
    catch (reason) { setError(getErrorMessage(reason)); setBusy(false); }
  }

  function reset() { setAnalysis(null); setFeasibility(null); setAnswers({}); setQuestionIndex(0); setError(""); }

  if (feasibility) {
    const allowed = feasibility.canGeneratePlan;
    return <section className="goal-composer result-panel"><span className="eyebrow">Оценка цели</span><h1>{allowed ? "Цель реалистична" : "Цель стоит скорректировать"}</h1><p>{feasibility.reason}</p>
      {feasibility.suggestedGoal ? <div className="suggestion"><strong>Рекомендация AI</strong><p>{feasibility.suggestedGoal.description}</p></div> : null}
      {error ? <div className="inline-error">{error}</div> : null}
      <div className="composer-actions">{allowed ? <button className="button button--primary" disabled={busy} onClick={generate}>{busy ? "Строим план..." : "Построить план"}</button> : feasibility.suggestedGoal ? <button className="button button--primary" disabled={busy} onClick={() => { const suggested = feasibility.suggestedGoal!; setText(suggested.description); void analyze(suggested.description); }}>Принять корректировку</button> : null}<button className="button button--ghost" onClick={reset}>Отменить</button></div>
    </section>;
  }

  if (analysis && question) return <section className="goal-composer question-panel"><div className="question-progress"><span>Вопрос {questionIndex + 1} из {questions.length}</span><i><b style={{width: `${progress}%`}} /></i></div><h1>{question.text}</h1><QuestionInput question={question} value={answers[question.id]} onChange={(value) => setAnswers({...answers, [question.id]: value})} />
    {error ? <div className="inline-error">{error}</div> : null}<div className="composer-actions"><button className="button button--ghost" disabled={questionIndex === 0 || busy} onClick={() => setQuestionIndex((value) => value - 1)}>Назад</button><button className="button button--primary" disabled={!canContinue || busy} onClick={continueQuestions}>{busy ? "Проверяем..." : questionIndex === questions.length - 1 ? "Оценить цель" : "Далее"}</button></div></section>;

  return <section className="goal-composer"><span className="eyebrow">Новая цель</span><h1>Что вы хотите изменить?</h1><p>Опишите цель своими словами. AI уточнит детали, проверит реалистичность и построит план.</p><form onSubmit={start}><textarea maxLength={2000} value={text} onChange={(e) => setText(e.target.value)} placeholder="Например: хочу изучить основы Go за три месяца" autoFocus /><div className="composer-footer"><span>{text.length} / 2000</span><button className="button button--primary" disabled={busy || text.trim().length < 5}>{busy ? "Анализируем..." : "Продолжить"}</button></div></form>{error ? <div className="inline-error">{error}</div> : null}</section>;
}

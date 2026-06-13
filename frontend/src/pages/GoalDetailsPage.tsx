import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { analyticsApi, decompositionApi, dependenciesApi, goalsApi, tasksApi } from "../api/endpoints";
import type { Goal, ProgressResponse, TaskDependency, TaskTreeNode } from "../api/types";
import { TaskTree } from "../components/TaskTree";
import { getErrorMessage } from "../utils/errors";
import { formatDate } from "../utils/format";

function flatten(nodes: TaskTreeNode[]): TaskTreeNode[] { return nodes.flatMap((node) => [node, ...flatten(node.children)]); }

export function GoalDetailsPage() {
  const { goalId } = useParams<{goalId: string}>();
  const navigate = useNavigate();
  const [goal, setGoal] = useState<Goal | null>(null);
  const [tasks, setTasks] = useState<TaskTreeNode[]>([]);
  const [progress, setProgress] = useState<ProgressResponse | null>(null);
  const [dependencies, setDependencies] = useState<Record<string, TaskDependency[]>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [regenerating, setRegenerating] = useState(false);

  async function load() {
    if (!goalId) return;
    setLoading(true); setError("");
    try {
      const [goalData, taskData, progressData] = await Promise.all([goalsApi.get(goalId), tasksApi.listByGoal(goalId), analyticsApi.getProgress(goalId)]);
      const depEntries = await Promise.all(flatten(taskData).map(async (task) => [task.id, await dependenciesApi.list(task.id)] as const));
      setGoal(goalData); setTasks(taskData); setProgress(progressData); setDependencies(Object.fromEntries(depEntries));
    } catch (reason) { setError(getErrorMessage(reason)); }
    finally { setLoading(false); }
  }

  useEffect(() => { void load(); }, [goalId]);

  async function removeGoal() {
    if (!goalId || !window.confirm("Удалить цель и весь её план? Это действие нельзя отменить после закрытия страницы.")) return;
    await goalsApi.remove(goalId); navigate("/goals");
  }

  async function regenerate() {
    if (!goalId || !window.confirm("Текущий план и весь прогресс будут удалены. Продолжить?")) return;
    setRegenerating(true); setError("");
    try { await decompositionApi.regenerateGoal(goalId, {detailLevel: "HIGH", includeDependencies: true, includeDeadlines: true}); await load(); }
    catch (reason) { setError(getErrorMessage(reason)); }
    finally { setRegenerating(false); }
  }

  if (loading) return <div className="empty-page">Загружаем план...</div>;
  if (!goal || !progress) return <div className="empty-page"><h2>Цель недоступна</h2><p>{error}</p></div>;

  return <section className="plan-page">
    <header className="plan-header"><div><button className="back-link" onClick={() => navigate("/goals")}>← Мои цели</button><h1>{goal.title}</h1><p>{goal.description}</p></div><div className="page-actions"><button className="button button--ghost" disabled={regenerating} onClick={regenerate}>{regenerating ? "Перестраиваем..." : "Перестроить план"}</button><button className="button button--danger" onClick={removeGoal}>Удалить цель</button></div></header>
    <div className="plan-summary"><div><span>Общий прогресс</span><strong>{Math.round(progress.simpleProgressPercent)}%</strong></div><div className="progress-line"><i style={{width: `${progress.simpleProgressPercent}%`}} /></div><span>Срок цели: {formatDate(goal.deadline)}</span></div>
    {error ? <div className="inline-error">{error}</div> : null}
    <TaskTree tasks={tasks} dependencies={dependencies} onChanged={load} />
  </section>;
}

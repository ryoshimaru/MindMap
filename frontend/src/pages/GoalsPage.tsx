import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { analyticsApi, goalsApi } from "../api/endpoints";
import type { Goal } from "../api/types";
import { getErrorMessage } from "../utils/errors";
import { formatDate } from "../utils/format";

type GoalRow = Goal & { progress: number };

export function GoalsPage() {
  const [goals, setGoals] = useState<GoalRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    goalsApi.list().then(async (items) => {
      const sorted = [...items].sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt));
      setGoals(await Promise.all(sorted.map(async (goal) => {
        const progress = await analyticsApi.getProgress(goal.id).catch(() => null);
        return {...goal, progress: progress?.simpleProgressPercent || 0};
      })));
    }).catch((reason) => setError(getErrorMessage(reason))).finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="empty-page">Загружаем цели...</div>;
  return <section className="goals-page"><div className="section-heading"><span className="eyebrow">История</span><h1>Мои цели</h1><p>Новые цели находятся сверху.</p></div>
    {error ? <div className="inline-error">{error}</div> : null}
    {goals.length === 0 ? <div className="empty-page"><h2>Здесь пока пусто</h2><p>Создайте первую цель, и её план появится здесь.</p><Link className="button button--primary" to="/dashboard">Создать цель</Link></div> :
      <div className="goal-list">{goals.map((goal) => <Link key={goal.id} to={`/goals/${goal.id}`} className="goal-row">
        <div><span className="goal-row__state">{goal.progress === 100 ? "Завершена" : "Активна"}</span><h2>{goal.title}</h2><p>Срок: {formatDate(goal.deadline)}</p></div>
        <div className="goal-row__progress"><strong>{Math.round(goal.progress)}%</strong><span><i style={{width: `${goal.progress}%`}} /></span></div>
      </Link>)}</div>}
  </section>;
}

import { Link } from "react-router-dom";
import type { Goal } from "../api/types";
import { formatDate } from "../utils/format";
import { StatusBadge } from "./StatusBadge";

interface GoalCardProps {
  goal: Goal;
  progressPercent: number;
}

export function GoalCard({ goal, progressPercent }: GoalCardProps) {
  return (
    <article className="goal-card card">
      <div className="goal-card__top">
        <div>
          <span className="eyebrow">{goal.category}</span>
          <strong>{goal.title}</strong>
        </div>
        <div className="page-header__meta">
          <StatusBadge value={goal.priority} tone="priority" />
          <StatusBadge value={goal.status} />
        </div>
      </div>

      <p>{goal.description}</p>

      <div className="goal-card__progress">
        <div className="card-header">
          <span className="metric-label">Прогресс</span>
          <strong>{progressPercent}%</strong>
        </div>
        <div className="progress-bar" aria-label={`Прогресс ${progressPercent}%`}>
          <span style={{ width: `${progressPercent}%` }} />
        </div>
      </div>

      <div className="goal-card__bottom">
        <div>
          <span className="metric-label">Срок</span>
          <strong>{formatDate(goal.deadline)}</strong>
        </div>
        <div>
          <span className="metric-label">Нагрузка</span>
          <strong>{goal.availableHoursPerWeek} ч/нед</strong>
        </div>
        <Link to={`/goals/${goal.id}`} className="button button--primary">
          Открыть
        </Link>
      </div>
    </article>
  );
}

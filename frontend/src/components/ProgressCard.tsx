import type { ProgressResponse } from "../api/types";

interface ProgressCardProps {
  progress: ProgressResponse;
}

export function ProgressCard({ progress }: ProgressCardProps) {
  return (
    <section className="card card--compact">
      <span className="eyebrow">Прогресс</span>
      <div className="progress-value">{progress.weightedProgressPercent}%</div>
      <div className="progress-bar" aria-label="Взвешенный прогресс">
        <span style={{ width: `${progress.weightedProgressPercent}%` }} />
      </div>
      <div className="detail-list">
        <div>
          <dt>Готово задач</dt>
          <dd>
            {progress.completedTasks} / {progress.totalTasks}
          </dd>
        </div>
        <div>
          <dt>Закрыто часов</dt>
          <dd>
            {progress.completedEstimatedHours} / {progress.totalEstimatedHours}
          </dd>
        </div>
      </div>
    </section>
  );
}

import type { MetricsResponse } from "../api/types";

interface MetricsCardProps {
  metrics: MetricsResponse;
}

export function MetricsCard({ metrics }: MetricsCardProps) {
  return (
    <section className="card card--compact">
      <span className="eyebrow">Метрики работы</span>
      <div className="metric-grid">
        <div className="metric-card">
          <span className="metric-label">Задач от AI</span>
          <strong className="metric-value">{metrics.tasksCreatedByAI}</strong>
        </div>
        <div className="metric-card">
          <span className="metric-label">Изменено вами</span>
          <strong className="metric-value">{metrics.tasksEditedByUser}</strong>
        </div>
        <div className="metric-card">
          <span className="metric-label">Удалено вами</span>
          <strong className="metric-value">{metrics.tasksDeletedByUser}</strong>
        </div>
        <div className="metric-card">
          <span className="metric-label">Завершено</span>
          <strong className="metric-value">{metrics.tasksCompleted}</strong>
        </div>
        <div className="metric-card">
          <span className="metric-label">Регенераций</span>
          <strong className="metric-value">{metrics.regenerationCount}</strong>
        </div>
        <div className="metric-card">
          <span className="metric-label">Принято задач</span>
          <strong className="metric-value">{metrics.acceptedTasksPercent}%</strong>
        </div>
      </div>
    </section>
  );
}

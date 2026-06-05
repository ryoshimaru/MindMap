import type { FeasibilityResponse } from "../api/types";
import { StatusBadge } from "./StatusBadge";

interface FeasibilityCardProps {
  feasibility: FeasibilityResponse;
}

export function FeasibilityCard({ feasibility }: FeasibilityCardProps) {
  return (
    <section className="card card--compact">
      <div className="card-header">
        <span className="eyebrow">Реалистичность</span>
        <StatusBadge value={feasibility.status} tone="feasibility" />
      </div>
      <p>{feasibility.message}</p>
      <div className="metric-grid">
        <div className="metric-card">
          <span className="metric-label">Недель осталось</span>
          <strong className="metric-value">{feasibility.weeksUntilDeadline}</strong>
        </div>
        <div className="metric-card">
          <span className="metric-label">Доступно часов</span>
          <strong className="metric-value">{feasibility.availableHours}</strong>
        </div>
        <div className="metric-card">
          <span className="metric-label">Нужно часов</span>
          <strong className="metric-value">{feasibility.requiredHours}</strong>
        </div>
      </div>
    </section>
  );
}

import type { QualityResponse } from "../api/types";

interface QualityScoreCardProps {
  quality: QualityResponse;
}

export function QualityScoreCard({ quality }: QualityScoreCardProps) {
  return (
    <section className="card card--compact">
      <span className="eyebrow">Качество плана</span>
      <div className="quality-score">{quality.score}/100</div>
      <div className="progress-bar" aria-label="Оценка качества">
        <span style={{ width: `${quality.score}%` }} />
      </div>
      <div className="section-stack">
        {quality.details.map((detail) => (
          <div key={detail} className="metric-card">
            {detail}
          </div>
        ))}
      </div>
    </section>
  );
}

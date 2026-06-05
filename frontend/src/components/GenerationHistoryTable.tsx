import type { Generation } from "../api/types";
import { formatDateTime } from "../utils/format";
import { StatusBadge } from "./StatusBadge";

interface GenerationHistoryTableProps {
  generations: Generation[];
  selectedGenerationId?: string | null;
  onSelectGeneration: (generationId: string) => void;
}

export function GenerationHistoryTable({
  generations,
  selectedGenerationId,
  onSelectGeneration
}: GenerationHistoryTableProps) {
  return (
    <div className="card">
      <div className="page-section__header">
        <div>
          <span className="eyebrow">Журнал генераций</span>
          <h2>Запуски AI-декомпозиции</h2>
        </div>
      </div>

      <table className="table">
        <thead>
          <tr>
            <th>Статус</th>
            <th>Модель</th>
            <th>Создано</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {generations.map((generation) => (
            <tr
              key={generation.id}
              className={
                generation.id === selectedGenerationId ? "table-row--active" : undefined
              }
            >
              <td>
                <StatusBadge value={generation.status} />
              </td>
              <td>{generation.modelName}</td>
              <td>{formatDateTime(generation.createdAt)}</td>
              <td>
                <button
                  type="button"
                  className="button button--ghost"
                  onClick={() => onSelectGeneration(generation.id)}
                >
                  Подробнее
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

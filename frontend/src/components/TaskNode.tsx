import type { TaskStatus, TaskTreeNode } from "../api/types";
import { formatDate } from "../utils/format";
import { StatusBadge } from "./StatusBadge";

interface TaskNodeProps {
  node: TaskTreeNode;
  depth: number;
  updatingTaskId?: string | null;
  expandingTaskId?: string | null;
  onStatusChange: (taskId: string, status: TaskStatus) => Promise<void>;
  onDecomposeTask?: (taskId: string) => Promise<void>;
}

export function TaskNode({
  node,
  depth,
  updatingTaskId,
  expandingTaskId,
  onStatusChange,
  onDecomposeTask
}: TaskNodeProps) {
  const levelLabel = depth === 0 ? "Этап" : depth === 1 ? "Задача" : "Подзадача";

  return (
    <article className="task-node">
      <div className="task-node__header">
        <div className="task-node__title">
          <span className="eyebrow">{levelLabel}</span>
          <strong>{node.title}</strong>
          <p>{node.description}</p>
        </div>

        <div className="page-header__meta">
          <StatusBadge value={node.priority} tone="priority" />
          <StatusBadge value={node.source} tone="source" />
        </div>
      </div>

      <div className="task-node__meta">
        <span>Статус</span>
        <select
          className="field-control"
          value={node.status}
          onChange={(event) =>
            void onStatusChange(node.id, event.target.value as TaskStatus)
          }
          disabled={updatingTaskId === node.id}
        >
          <option value="TODO">К выполнению</option>
          <option value="IN_PROGRESS">В работе</option>
          <option value="DONE">Готово</option>
          <option value="CANCELLED">Отменено</option>
        </select>
        <StatusBadge value={node.status} />
        <span>{node.estimatedHours} ч</span>
        <span>Срок: {formatDate(node.deadline)}</span>
        {node.editedByUser ? <span>Изменено пользователем</span> : null}
      </div>

      {onDecomposeTask ? (
        <div className="inline-form-actions">
          <button
            type="button"
            className="button button--ghost"
            disabled={expandingTaskId === node.id}
            onClick={() => void onDecomposeTask(node.id)}
          >
            {expandingTaskId === node.id ? "Детализация..." : "Детализировать с AI"}
          </button>
        </div>
      ) : null}

      {node.children.length > 0 ? (
        <div className="task-children">
          {node.children.map((childNode) => (
            <TaskNode
              key={childNode.id}
              node={childNode}
              depth={depth + 1}
              updatingTaskId={updatingTaskId}
              expandingTaskId={expandingTaskId}
              onStatusChange={onStatusChange}
              onDecomposeTask={onDecomposeTask}
            />
          ))}
        </div>
      ) : null}
    </article>
  );
}

import type { TaskStatus, TaskTreeNode } from "../api/types";
import { EmptyState } from "./EmptyState";
import { TaskNode } from "./TaskNode";

interface TaskTreeProps {
  tasks: TaskTreeNode[];
  updatingTaskId?: string | null;
  expandingTaskId?: string | null;
  onStatusChange: (taskId: string, status: TaskStatus) => Promise<void>;
  onDecomposeTask?: (taskId: string) => Promise<void>;
}

export function TaskTree({
  tasks,
  updatingTaskId,
  expandingTaskId,
  onStatusChange,
  onDecomposeTask
}: TaskTreeProps) {
  if (tasks.length === 0) {
    return (
      <EmptyState
        title="Плана задач пока нет"
        description="Постройте AI-план, чтобы превратить цель в этапы, задачи и подзадачи."
      />
    );
  }

  return (
    <div className="task-tree">
      {tasks.map((task) => (
        <TaskNode
          key={task.id}
          node={task}
          depth={0}
          updatingTaskId={updatingTaskId}
          expandingTaskId={expandingTaskId}
          onStatusChange={onStatusChange}
          onDecomposeTask={onDecomposeTask}
        />
      ))}
    </div>
  );
}

import { useState } from "react";
import { tasksApi } from "../api/endpoints";
import type { Task, TaskDependency, TaskTreeNode } from "../api/types";
import { TaskNode } from "./TaskNode";

function flatten(nodes: TaskTreeNode[]): TaskTreeNode[] { return nodes.flatMap((node) => [node, ...flatten(node.children)]); }

export function TaskTree({ tasks, dependencies, onChanged, onStatusChanged, onBeforeStatusChange }: { tasks: TaskTreeNode[]; dependencies: Record<string, TaskDependency[]>; onChanged: () => Promise<void>; onStatusChanged: (task: Task) => Promise<void>; onBeforeStatusChange: () => void }) {
  const [addingStage, setAddingStage] = useState(false);
  const [title, setTitle] = useState("");
  const allTasks = flatten(tasks);

  async function addStage() {
    if (!title.trim() || !tasks[0]) return;
    await tasksApi.create(tasks[0].goalId, {title, description: "", priority: "MEDIUM", estimatedHours: 0, orderIndex: tasks.length + 1});
    setTitle(""); setAddingStage(false); await onChanged();
  }

  return <div className="task-tree">
    {tasks.map((stage, index) => <TaskNode key={stage.id} node={stage} stageIndex={index} siblings={tasks} allTasks={allTasks} dependencies={dependencies} onChanged={onChanged} onStatusChanged={onStatusChanged} onBeforeStatusChange={onBeforeStatusChange} />)}
    {addingStage ? <div className="inline-editor"><input autoFocus value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Название этапа"/><button className="button button--primary" onClick={addStage}>Добавить</button><button className="button button--ghost" onClick={() => setAddingStage(false)}>Отмена</button></div> : <button className="add-stage" onClick={() => setAddingStage(true)}>+ Добавить этап</button>}
  </div>;
}

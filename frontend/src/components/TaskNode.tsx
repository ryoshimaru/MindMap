import { useState } from "react";
import { dependenciesApi, tasksApi } from "../api/endpoints";
import type { Task, TaskDependency, TaskTreeNode } from "../api/types";
import { formatDate } from "../utils/format";

export function TaskNode({ node, stageIndex, siblings, allTasks, dependencies, onChanged, onStatusChanged, onBeforeStatusChange }: { node: TaskTreeNode; stageIndex: number; siblings: TaskTreeNode[]; allTasks: TaskTreeNode[]; dependencies: Record<string, TaskDependency[]>; onChanged: () => Promise<void>; onStatusChanged: (task: Task) => Promise<void>; onBeforeStatusChange: () => void }) {
  const [editing, setEditing] = useState(false);
  const [adding, setAdding] = useState(false);
  const [form, setForm] = useState({title: node.title, description: node.description, deadline: node.deadline || ""});
  const [newTitle, setNewTitle] = useState("");
  const [newDeadline, setNewDeadline] = useState("");
  const deps = dependencies[node.id] || [];
  const locked = deps.some((dep) => allTasks.find((task) => task.id === dep.dependsOnTaskId)?.status !== "DONE");
  const isStage = node.parentTaskId === null;
  const overdue = !isStage && node.status !== "DONE" && node.deadline && node.deadline < new Date().toISOString().slice(0, 10);
  const siblingIndex = siblings.findIndex((item) => item.id === node.id);

  async function save() { await tasksApi.update(node.id, {title: form.title, description: form.description, deadline: form.deadline || null}); setEditing(false); await onChanged(); }
  async function toggle() {
    onBeforeStatusChange();
    const updated = await tasksApi.updateStatus(node.id, {status: node.status === "DONE" ? "TODO" : "DONE"});
    await onStatusChanged(updated);
  }
  async function remove() { if (window.confirm(`Удалить «${node.title}»${isStage ? " вместе со всеми шагами" : ""}?`)) { await tasksApi.remove(node.id); await onChanged(); } }
  async function addStep() { if (!newTitle.trim()) return; await tasksApi.create(node.goalId, {parentTaskId: node.id, title: newTitle, description: "", deadline: newDeadline || null, priority: "MEDIUM", estimatedHours: 0, orderIndex: node.children.length + 1}); setNewTitle(""); setNewDeadline(""); setAdding(false); await onChanged(); }
  async function setDependency(dependsOnTaskId: string) {
    await Promise.all(deps.map((dep) => dependenciesApi.remove(node.id, dep.id)));
    if (dependsOnTaskId) await dependenciesApi.create(node.id, {dependsOnTaskId});
    await onChanged();
  }
  async function move(offset: number) {
    const target = siblings[siblingIndex + offset];
    if (!target) return;
    await Promise.all([tasksApi.update(node.id, {orderIndex: target.orderIndex}), tasksApi.update(target.id, {orderIndex: node.orderIndex})]);
    await onChanged();
  }

  if (!isStage) return <article className={`step-row${locked ? " step-row--locked" : ""}${overdue ? " step-row--overdue" : ""}`}>
    <button className={`step-check${node.status === "DONE" ? " step-check--done" : ""}`} disabled={locked} onClick={toggle} aria-pressed={node.status === "DONE"} aria-label={node.status === "DONE" ? "Вернуть шаг" : "Выполнить шаг"}>{locked ? "⌑" : null}</button>
    {editing ? <div className="step-editor"><input value={form.title} onChange={(e) => setForm({...form, title: e.target.value})}/><textarea value={form.description} onChange={(e) => setForm({...form, description: e.target.value})}/><input type="date" value={form.deadline} onChange={(e) => setForm({...form, deadline: e.target.value})}/><select value={node.parentTaskId || ""} onChange={(e) => void tasksApi.update(node.id, {parentTaskId: e.target.value}).then(onChanged)}>{allTasks.filter((task) => task.parentTaskId === null).map((stage) => <option key={stage.id} value={stage.id}>Этап: {stage.title}</option>)}</select><select value={deps[0]?.dependsOnTaskId || ""} onChange={(e) => void setDependency(e.target.value)}><option value="">Без зависимости</option>{allTasks.filter((task) => task.parentTaskId && task.id !== node.id).map((task) => <option key={task.id} value={task.id}>После: {task.title}</option>)}</select><div><button className="button button--primary" onClick={save}>Сохранить</button><button className="button button--ghost" onClick={() => setEditing(false)}>Отмена</button></div></div> : <><div className="step-copy"><div><h3>{node.title}</h3>{locked ? <span className="lock-label">Заблокирован предыдущим шагом</span> : null}</div><p>{node.description}</p><span className={overdue ? "deadline deadline--overdue" : "deadline"}>{overdue ? "Просрочено: " : "Дедлайн: "}{formatDate(node.deadline)}</span></div><div className="node-actions"><button disabled={siblingIndex === 0} onClick={() => void move(-1)}>↑</button><button disabled={siblingIndex === siblings.length - 1} onClick={() => void move(1)}>↓</button><button onClick={() => setEditing(true)}>Изменить</button><button onClick={remove}>Удалить</button></div></>}
  </article>;

  return <section className="stage-card"><header><div><span className="stage-number">{String(stageIndex + 1).padStart(2, "0")}</span><h2>{node.title}</h2></div><div className="node-actions"><button disabled={siblingIndex === 0} onClick={() => void move(-1)}>↑</button><button disabled={siblingIndex === siblings.length - 1} onClick={() => void move(1)}>↓</button><button onClick={() => setEditing(!editing)}>Изменить</button><button onClick={remove}>Удалить</button></div></header>
    {editing ? <div className="inline-editor"><input value={form.title} onChange={(e) => setForm({...form, title: e.target.value})}/><button className="button button--primary" onClick={save}>Сохранить</button></div> : null}
    <div className="stage-steps">{node.children.map((child) => <TaskNode key={child.id} node={child} stageIndex={stageIndex} siblings={node.children} allTasks={allTasks} dependencies={dependencies} onChanged={onChanged} onStatusChanged={onStatusChanged} onBeforeStatusChange={onBeforeStatusChange}/>)}</div>
    {adding ? <div className="inline-editor"><input autoFocus value={newTitle} onChange={(e) => setNewTitle(e.target.value)} placeholder="Название шага"/><input type="date" value={newDeadline} onChange={(e) => setNewDeadline(e.target.value)}/><button className="button button--primary" onClick={addStep}>Добавить</button><button className="button button--ghost" onClick={() => setAdding(false)}>Отмена</button></div> : <button className="add-step" onClick={() => setAdding(true)}>+ Добавить шаг</button>}
  </section>;
}

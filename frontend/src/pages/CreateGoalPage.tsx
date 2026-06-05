import { SmartGoalRequestPanel } from "../components/SmartGoalRequestPanel";

export function CreateGoalPage() {
  return (
    <div className="dashboard-workspace">
      <SmartGoalRequestPanel />

      <aside className="card">
        <span className="eyebrow">Новый сценарий</span>
        <h3>Больше не нужно заполнять поля вручную</h3>
        <ol className="api-checklist">
          <li>Введите цель естественным языком.</li>
          <li>Проверьте интерпретацию AI.</li>
          <li>Ответьте на уточняющие вопросы.</li>
          <li>Постройте план и перейдите к дереву задач.</li>
        </ol>
      </aside>
    </div>
  );
}

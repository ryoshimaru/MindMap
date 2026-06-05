export function ApiDocsPage() {
  return (
    <div className="section-stack">
      <section className="card docs-banner">
        <div>
          <span className="eyebrow">Backend-контракт</span>
          <h2>Интеграция frontend по OpenAPI</h2>
          <p>
            Frontend готов работать с backend, который реализует контракт из
            <span className="text-code">docs/openapi.yaml</span>.
          </p>
        </div>
        <div className="page-actions">
          <span className="pill pill--accent">OpenAPI 3.1</span>
          <span className="pill">Bearer JWT</span>
          <span className="pill">Демо-клиент</span>
        </div>
      </section>

      <section className="docs-grid">
        <article className="card docs-card">
          <span className="eyebrow">Аутентификация</span>
          <h3>Вход и сессия</h3>
          <p>
            В MVP JWT хранится в <span className="text-code">localStorage</span>.
            Клиент автоматически добавляет
            <span className="text-code">Authorization: Bearer &lt;token&gt;</span>.
          </p>
          <ul>
            <li>POST /api/auth/register</li>
            <li>POST /api/auth/login</li>
            <li>GET /api/auth/me</li>
            <li>Google OAuth endpoints описаны для redirect-сценария.</li>
          </ul>
        </article>

        <article className="card docs-card">
          <span className="eyebrow">Планирование</span>
          <h3>Цели, задачи и версии</h3>
          <p>
            Контракт покрывает CRUD целей, контекст, дерево задач, зависимости, версии
            плана и историю генераций.
          </p>
          <ul>
            <li>Цели и контекст находятся в /api/goals</li>
            <li>Задачи поддерживают CRUD и PATCH /status</li>
            <li>Версии плана можно активировать без удаления истории</li>
          </ul>
        </article>

        <article className="card docs-card">
          <span className="eyebrow">AI и аналитика</span>
          <h3>Декомпозиция и оценка</h3>
          <p>
            GoalMind использует AI endpoints вместе с прогрессом, реалистичностью,
            качеством и метриками использования.
          </p>
          <ul>
            <li>POST /api/requests/analyze</li>
            <li>POST /api/requests/{`{requestId}`}/answers</li>
            <li>POST /api/requests/{`{requestId}`}/generate-plan</li>
            <li>GET /api/goals/{`{goalId}`}/progress, /feasibility, /quality, /metrics</li>
          </ul>
        </article>
      </section>

      <section className="card">
        <span className="eyebrow">Файлы проекта</span>
        <h3>Где смотреть контракт</h3>
        <ul className="api-checklist">
          <li>
            Основная OpenAPI-спецификация:
            <span className="text-code">docs/openapi.yaml</span>
          </li>
          <li>
            Backend-справка:
            <span className="text-code">docs/api-overview.md</span>
          </li>
          <li>
            Frontend API-клиент:
            <span className="text-code">frontend/src/api/endpoints.ts</span>
          </li>
          <li>
            Демо-данные:
            <span className="text-code">frontend/src/api/mockData.ts</span>
          </li>
        </ul>
      </section>
    </div>
  );
}

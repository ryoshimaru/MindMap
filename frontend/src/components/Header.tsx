import { Link, useLocation } from "react-router-dom";
import { USE_MOCKS } from "../api/client";
import { useAuth } from "../auth/useAuth";

const routeCopy: Array<{
  matcher: (pathname: string) => boolean;
  title: string;
  description: string;
}> = [
  {
    matcher: (pathname) => pathname.startsWith("/goals/") && pathname.includes("/history"),
    title: "История генераций",
    description: "AI-запуски, промпт, ответ модели и обратная связь."
  },
  {
    matcher: (pathname) => pathname.startsWith("/goals/"),
    title: "Roadmap",
    description: "Дерево задач, статусы, прогресс и версии плана."
  },
  {
    matcher: (pathname) => pathname.startsWith("/api-docs"),
    title: "Контракт API",
    description: "Справка по backend-контракту."
  },
  {
    matcher: () => true,
    title: "GoalMind",
    description: "Рабочее пространство цели."
  }
];

export function Header() {
  const location = useLocation();
  const { logout } = useAuth();
  const headerCopy =
    routeCopy.find((item) => item.matcher(location.pathname)) || routeCopy[0];

  return (
    <header className="page-header">
      <div>
        <div className="page-header__meta">
          <span className="eyebrow">GoalMind</span>
          {USE_MOCKS ? <span className="pill pill--accent">Демо-режим</span> : null}
        </div>
        <h1>{headerCopy.title}</h1>
        <p>{headerCopy.description}</p>
      </div>

      <div className="page-header__actions">
        <Link to="/dashboard" className="button button--ghost">
          Новая цель
        </Link>
        <button type="button" className="button button--secondary" onClick={logout}>
          Выйти
        </button>
      </div>
    </header>
  );
}

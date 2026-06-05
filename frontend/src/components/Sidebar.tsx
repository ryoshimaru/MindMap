import { NavLink } from "react-router-dom";

const navigationItems = [{ to: "/dashboard", label: "Новая цель" }];

export function Sidebar() {
  return (
    <aside className="sidebar">
      <div className="brand-mark">
        <span className="brand-mark__eyebrow">GoalMind</span>
        <strong>Roadmap</strong>
      </div>

      <nav className="sidebar-nav" aria-label="Основная навигация">
        {navigationItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              `sidebar-link${isActive ? " sidebar-link--active" : ""}`
            }
          >
            {item.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}

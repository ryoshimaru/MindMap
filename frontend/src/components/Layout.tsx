import { Outlet, useLocation } from "react-router-dom";
import { Header } from "./Header";
import { Sidebar } from "./Sidebar";

export function Layout() {
  const location = useLocation();
  const isDashboard = location.pathname === "/dashboard";

  return (
    <div className={`app-shell${isDashboard ? " app-shell--focus" : ""}`}>
      {isDashboard ? null : <Sidebar />}
      <div className="app-main">
        {isDashboard ? null : <Header />}
        <main className={isDashboard ? "app-content app-content--focus" : "app-content"}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}

import { Navigate, Outlet, useLocation } from "react-router-dom";
import { LoadingState } from "../components/LoadingState";
import { useAuth } from "./useAuth";

export function ProtectedRoute() {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <LoadingState
        title="Восстанавливаем рабочее пространство"
        description="Проверяем сессию и загружаем последние цели."
        fullPage
      />
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <Outlet />;
}

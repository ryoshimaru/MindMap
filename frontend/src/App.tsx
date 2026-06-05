import { Navigate, Route, Routes } from "react-router-dom";
import { ProtectedRoute } from "./auth/ProtectedRoute";
import { Layout } from "./components/Layout";
import { ApiDocsPage } from "./pages/ApiDocsPage";
import { CreateGoalPage } from "./pages/CreateGoalPage";
import { DashboardPage } from "./pages/DashboardPage";
import { GenerationHistoryPage } from "./pages/GenerationHistoryPage";
import { GoalDetailsPage } from "./pages/GoalDetailsPage";
import { LoginPage } from "./pages/LoginPage";
import { OAuthCallbackPage } from "./pages/OAuthCallbackPage";
import { RegisterPage } from "./pages/RegisterPage";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/auth/callback" element={<OAuthCallbackPage />} />

      <Route element={<ProtectedRoute />}>
        <Route element={<Layout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/goals/new" element={<CreateGoalPage />} />
          <Route path="/goals/:goalId" element={<GoalDetailsPage />} />
          <Route
            path="/goals/:goalId/history"
            element={<GenerationHistoryPage />}
          />
          <Route path="/api-docs" element={<ApiDocsPage />} />
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}

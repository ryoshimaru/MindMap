import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../auth/useAuth";
import { ErrorState } from "../components/ErrorState";
import { LoadingState } from "../components/LoadingState";

export function OAuthCallbackPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { authenticateWithToken } = useAuth();
  const [errorMessage, setErrorMessage] = useState("");

  useEffect(() => {
    const token = searchParams.get("token");

    if (!token) {
      setErrorMessage("OAuth-ответ не содержит токен.");
      return;
    }

    void (async () => {
      try {
        await authenticateWithToken(token);
        navigate("/dashboard", { replace: true });
      } catch (error) {
        setErrorMessage(
          error instanceof Error
            ? error.message
            : "Не удалось завершить вход через OAuth."
        );
      }
    })();
  }, [navigate, searchParams]);

  if (errorMessage) {
    return (
      <div className="auth-shell">
        <div className="auth-grid">
          <section className="auth-panel">
            <ErrorState
              title="Не удалось войти через OAuth"
              description={errorMessage}
              actionLabel="Вернуться ко входу"
              onAction={() => navigate("/login", { replace: true })}
            />
            <Link to="/login" className="link-button">
              Вернуться ко входу
            </Link>
          </section>
        </div>
      </div>
    );
  }

  return (
    <LoadingState
      title="Завершаем вход"
      description="mindmap проверяет OAuth-ответ и открывает рабочее пространство."
      fullPage
    />
  );
}

import { useState } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { USE_MOCKS } from "../api/client";
import { authApi } from "../api/endpoints";
import { useAuth } from "../auth/useAuth";
import { getErrorMessage } from "../utils/errors";

export function LoginPage() {
  const { isAuthenticated, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  const fromPath =
    ((location.state as { from?: { pathname?: string } } | null)?.from?.pathname ??
      "/dashboard");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setErrorMessage("");

    try {
      await login({ email, password });
      navigate(fromPath, { replace: true });
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  }

  function handleGoogleLogin() {
    window.location.href = authApi.getGoogleStartUrl();
  }

  return (
    <div className="auth-shell">
      <div className="auth-grid">
        <section className="auth-brand">
          <span className="eyebrow">GoalMind</span>
          <strong>Вход в рабочее пространство.</strong>
          <p>
            После входа вы сразу попадете к главному сценарию: свободный ввод цели,
            AI-анализ, уточнения и построение roadmap.
          </p>
        </section>

        <section className="auth-panel">
          <span className="eyebrow">Вход</span>
          <h1>Добро пожаловать</h1>
          <p>Введите почту и пароль, чтобы открыть главный экран.</p>

          <form className="form-grid" onSubmit={handleSubmit}>
            <div className="field-group">
              <label className="field-label" htmlFor="login-email">
                Почта
              </label>
              <input
                id="login-email"
                className="field-control"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                required
              />
            </div>

            <div className="field-group">
              <label className="field-label" htmlFor="login-password">
                Пароль
              </label>
              <input
                id="login-password"
                className="field-control"
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
              />
            </div>

            {errorMessage ? <div className="inline-error">{errorMessage}</div> : null}

            {USE_MOCKS ? (
              <div className="feedback-success">
                Демо-режим включен только для обычного входа. Google всегда открывает
                backend OAuth flow.
              </div>
            ) : null}

            <button type="submit" className="button button--primary" disabled={isSubmitting}>
              {isSubmitting ? "Вход..." : "Войти"}
            </button>
          </form>

          <div className="auth-divider">или</div>

          <button type="button" className="button button--secondary" onClick={handleGoogleLogin}>
            Войти через Google
          </button>

          <div className="auth-panel__footer">
            <span className="field-hint">Нет аккаунта?</span>
            <Link to="/register" className="link-button">
              Зарегистрироваться
            </Link>
          </div>
        </section>
      </div>
    </div>
  );
}

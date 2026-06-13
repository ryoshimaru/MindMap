import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/useAuth";
import { getErrorMessage } from "../utils/errors";

export function RegisterPage() {
  const { isAuthenticated, register } = useAuth();
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setErrorMessage("");

    try {
      await register({ name, email, password });
      navigate("/dashboard", { replace: true });
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="auth-shell">
      <div className="auth-grid">
        <section className="auth-brand">
          <span className="eyebrow">mindmap</span>
          <strong>Создание аккаунта.</strong>
          <p>
            После регистрации откроется главный сценарий: ввод цели, AI-анализ,
            уточняющие вопросы и план.
          </p>
        </section>

        <section className="auth-panel">
          <span className="eyebrow">Регистрация</span>
          <h1>Новый аккаунт</h1>
          <p>Заполните имя, почту и пароль.</p>

          <form className="form-grid" onSubmit={handleSubmit}>
            <div className="field-group">
              <label className="field-label" htmlFor="register-name">
                Имя
              </label>
              <input
                id="register-name"
                className="field-control"
                value={name}
                onChange={(event) => setName(event.target.value)}
                required
              />
            </div>

            <div className="field-group">
              <label className="field-label" htmlFor="register-email">
                Почта
              </label>
              <input
                id="register-email"
                className="field-control"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                required
              />
            </div>

            <div className="field-group">
              <label className="field-label" htmlFor="register-password">
                Пароль
              </label>
              <input
                id="register-password"
                className="field-control"
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
              />
            </div>

            {errorMessage ? <div className="inline-error">{errorMessage}</div> : null}

            <button type="submit" className="button button--primary" disabled={isSubmitting}>
              {isSubmitting ? "Создание..." : "Зарегистрироваться"}
            </button>
          </form>

          <div className="auth-panel__footer">
            <span className="field-hint">Уже есть аккаунт?</span>
            <Link to="/login" className="link-button">
              Войти
            </Link>
          </div>
        </section>
      </div>
    </div>
  );
}

import { useEffect, useState } from "react";
import type { GoalContextInput } from "../api/types";

interface GoalContextFormProps {
  initialValues: GoalContextInput;
  isSaving?: boolean;
  onSubmit: (payload: GoalContextInput) => Promise<void>;
}

export function GoalContextForm({
  initialValues,
  isSaving = false,
  onSubmit
}: GoalContextFormProps) {
  const [formValues, setFormValues] = useState({
    currentLevel: initialValues.currentLevel,
    constraintsText: initialValues.constraintsText,
    preferencesText: initialValues.preferences.join(", "),
    expectedResult: initialValues.expectedResult
  });

  useEffect(() => {
    setFormValues({
      currentLevel: initialValues.currentLevel,
      constraintsText: initialValues.constraintsText,
      preferencesText: initialValues.preferences.join(", "),
      expectedResult: initialValues.expectedResult
    });
  }, [initialValues]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onSubmit({
      currentLevel: formValues.currentLevel.trim(),
      constraintsText: formValues.constraintsText.trim(),
      preferences: formValues.preferencesText
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
      expectedResult: formValues.expectedResult.trim()
    });
  }

  return (
    <form className="form-grid" onSubmit={handleSubmit}>
      <div className="field-group">
        <label className="field-label" htmlFor="context-current-level">
          Текущий уровень
        </label>
        <textarea
          id="context-current-level"
          className="field-control"
          value={formValues.currentLevel}
          onChange={(event) =>
            setFormValues((currentValues) => ({
              ...currentValues,
              currentLevel: event.target.value
            }))
          }
          placeholder="Опишите, с какого состояния вы начинаете."
        />
      </div>

      <div className="field-group">
        <label className="field-label" htmlFor="context-constraints">
          Ограничения
        </label>
        <textarea
          id="context-constraints"
          className="field-control"
          value={formValues.constraintsText}
          onChange={(event) =>
            setFormValues((currentValues) => ({
              ...currentValues,
              constraintsText: event.target.value
            }))
          }
          placeholder="Укажите сроки, инструменты, ресурсы или другие ограничения."
        />
      </div>

      <div className="field-group">
        <label className="field-label" htmlFor="context-preferences">
          Предпочтения
        </label>
        <input
          id="context-preferences"
          className="field-control"
          value={formValues.preferencesText}
          onChange={(event) =>
            setFormValues((currentValues) => ({
              ...currentValues,
              preferencesText: event.target.value
            }))
          }
          placeholder="Понятный интерфейс, меньше зависимостей, сильные метрики"
        />
        <span className="field-hint">Разделяйте предпочтения запятыми.</span>
      </div>

      <div className="field-group">
        <label className="field-label" htmlFor="context-result">
          Ожидаемый результат
        </label>
        <textarea
          id="context-result"
          className="field-control"
          value={formValues.expectedResult}
          onChange={(event) =>
            setFormValues((currentValues) => ({
              ...currentValues,
              expectedResult: event.target.value
            }))
          }
          placeholder="Как должен выглядеть готовый результат?"
        />
      </div>

      <div className="inline-form-actions">
        <button type="submit" className="button button--secondary" disabled={isSaving}>
          {isSaving ? "Сохранение..." : "Сохранить контекст"}
        </button>
      </div>
    </form>
  );
}

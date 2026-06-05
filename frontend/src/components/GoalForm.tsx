import { useState } from "react";
import type { CreateGoalRequest, GoalPriority } from "../api/types";

interface GoalFormProps {
  initialValues?: CreateGoalRequest;
  isSubmitting?: boolean;
  submitLabel?: string;
  onSubmit: (payload: CreateGoalRequest) => Promise<void>;
}

const defaultValues: CreateGoalRequest = {
  title: "",
  description: "",
  category: "",
  priority: "MEDIUM",
  deadline: null,
  availableHoursPerWeek: 8
};

export function GoalForm({
  initialValues = defaultValues,
  isSubmitting = false,
  submitLabel = "Создать цель",
  onSubmit
}: GoalFormProps) {
  const [formValues, setFormValues] = useState<CreateGoalRequest>(initialValues);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onSubmit({
      ...formValues,
      title: formValues.title.trim(),
      description: formValues.description.trim(),
      category: formValues.category.trim(),
      deadline: formValues.deadline || null
    });
  }

  function updatePriority(priority: GoalPriority) {
    setFormValues((currentValues) => ({
      ...currentValues,
      priority
    }));
  }

  return (
    <form className="form-grid" onSubmit={handleSubmit}>
      <div className="field-group">
        <label className="field-label" htmlFor="goal-title">
          Название
        </label>
        <input
          id="goal-title"
          className="field-control"
          value={formValues.title}
          onChange={(event) =>
            setFormValues((currentValues) => ({
              ...currentValues,
              title: event.target.value
            }))
          }
          placeholder="Например: сформулировать цель"
          required
        />
      </div>

      <div className="field-group">
        <label className="field-label" htmlFor="goal-description">
          Описание
        </label>
        <textarea
          id="goal-description"
          className="field-control"
          value={formValues.description}
          onChange={(event) =>
            setFormValues((currentValues) => ({
              ...currentValues,
              description: event.target.value
            }))
          }
          placeholder="Коротко опишите ожидаемый результат и границы задачи."
          required
        />
      </div>

      <div className="form-grid form-grid--two">
        <div className="field-group">
          <label className="field-label" htmlFor="goal-category">
            Категория
          </label>
          <input
            id="goal-category"
            className="field-control"
            value={formValues.category}
            onChange={(event) =>
              setFormValues((currentValues) => ({
                ...currentValues,
                category: event.target.value
              }))
            }
            placeholder="Учеба, работа, продукт"
            required
          />
        </div>

        <div className="field-group">
          <label className="field-label" htmlFor="goal-deadline">
            Срок
          </label>
          <input
            id="goal-deadline"
            type="date"
            className="field-control"
            value={formValues.deadline ?? ""}
            onChange={(event) =>
              setFormValues((currentValues) => ({
                ...currentValues,
                deadline: event.target.value || null
              }))
            }
          />
        </div>
      </div>

      <div className="form-grid form-grid--two">
        <div className="field-group">
          <label className="field-label" htmlFor="goal-priority">
            Приоритет
          </label>
          <select
            id="goal-priority"
            className="field-control"
            value={formValues.priority}
            onChange={(event) => updatePriority(event.target.value as GoalPriority)}
          >
            <option value="LOW">Низкий</option>
            <option value="MEDIUM">Средний</option>
            <option value="HIGH">Высокий</option>
          </select>
        </div>

        <div className="field-group">
          <label className="field-label" htmlFor="goal-hours">
            Часов в неделю
          </label>
          <input
            id="goal-hours"
            type="number"
            min={1}
            step={1}
            className="field-control"
            value={formValues.availableHoursPerWeek}
            onChange={(event) =>
              setFormValues((currentValues) => ({
                ...currentValues,
                availableHoursPerWeek: Number(event.target.value) || 0
              }))
            }
            required
          />
        </div>
      </div>

      <div className="inline-form-actions">
        <button type="submit" className="button button--primary" disabled={isSubmitting}>
          {isSubmitting ? "Сохранение..." : submitLabel}
        </button>
        <span className="field-hint">
          Эта форма оставлена для совместимости со старым API. Основной сценарий теперь
          начинается с умного поля на главной странице.
        </span>
      </div>
    </form>
  );
}

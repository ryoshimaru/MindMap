export function formatDate(value: string | null) {
  if (!value) {
    return "Не указан";
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "short",
    year: "numeric"
  }).format(new Date(value));
}

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

export function humanizeEnum(value: string) {
  const labels: Record<string, string> = {
    DRAFT: "Черновик",
    ACTIVE: "Активно",
    COMPLETED: "Завершено",
    ARCHIVED: "В архиве",
    TODO: "К выполнению",
    IN_PROGRESS: "В работе",
    DONE: "Готово",
    CANCELLED: "Отменено",
    LOW: "Низкий",
    MEDIUM: "Средний",
    HIGH: "Высокий",
    REALISTIC: "Реалистично",
    RISKY: "Рискованно",
    UNREALISTIC: "Нереалистично",
    UNKNOWN: "Неизвестно",
    AI: "AI",
    MANUAL: "Вручную",
    PENDING: "Ожидает",
    SUCCESS: "Успешно",
    FAILED: "Ошибка"
  };

  return labels[value] || value;
}

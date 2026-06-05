interface ErrorStateProps {
  title?: string;
  description: string;
  actionLabel?: string;
  onAction?: () => void;
}

export function ErrorState({
  title = "Что-то пошло не так",
  description,
  actionLabel,
  onAction
}: ErrorStateProps) {
  return (
    <div className="error-state card">
      <span className="eyebrow">Нужно внимание</span>
      <h3>{title}</h3>
      <p>{description}</p>
      {actionLabel && onAction ? (
        <button type="button" className="button button--secondary" onClick={onAction}>
          {actionLabel}
        </button>
      ) : null}
    </div>
  );
}

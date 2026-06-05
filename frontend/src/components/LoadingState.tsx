interface LoadingStateProps {
  title?: string;
  description?: string;
  fullPage?: boolean;
}

export function LoadingState({
  title = "Загрузка",
  description = "GoalMind готовит следующий экран.",
  fullPage = false
}: LoadingStateProps) {
  return (
    <div className={fullPage ? "loading-state loading-state--full" : "loading-state"}>
      <div className="loading-spinner" aria-hidden="true" />
      <div>
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

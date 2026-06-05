import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { feedbackApi, generationsApi, goalsApi } from "../api/endpoints";
import type { Generation, GenerationDetails, Goal } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { ErrorState } from "../components/ErrorState";
import { FeedbackForm } from "../components/FeedbackForm";
import { GenerationHistoryTable } from "../components/GenerationHistoryTable";
import { LoadingState } from "../components/LoadingState";
import { StatusBadge } from "../components/StatusBadge";
import { getErrorMessage } from "../utils/errors";
import { formatDateTime } from "../utils/format";

export function GenerationHistoryPage() {
  const { goalId } = useParams<{ goalId: string }>();
  const [goal, setGoal] = useState<Goal | null>(null);
  const [generations, setGenerations] = useState<Generation[]>([]);
  const [selectedGenerationId, setSelectedGenerationId] = useState<string | null>(null);
  const [selectedGeneration, setSelectedGeneration] = useState<GenerationDetails | null>(
    null
  );
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmittingFeedback, setIsSubmittingFeedback] = useState(false);
  const [successMessage, setSuccessMessage] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  async function loadHistory() {
    if (!goalId) {
      setErrorMessage("В маршруте отсутствует идентификатор цели.");
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setErrorMessage("");

    try {
      const [goalResponse, generationList] = await Promise.all([
        goalsApi.get(goalId),
        generationsApi.listByGoal(goalId)
      ]);
      setGoal(goalResponse);
      setGenerations(generationList);

      const nextGenerationId = generationList[0]?.id ?? null;
      setSelectedGenerationId(nextGenerationId);

      if (nextGenerationId) {
        const details = await generationsApi.get(nextGenerationId);
        setSelectedGeneration(details);
      } else {
        setSelectedGeneration(null);
      }
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void loadHistory();
  }, [goalId]);

  async function handleSelectGeneration(generationId: string) {
    setSelectedGenerationId(generationId);
    setSuccessMessage("");

    try {
      const details = await generationsApi.get(generationId);
      setSelectedGeneration(details);
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    }
  }

  async function handleSubmitFeedback(payload: { rating: number; comment: string }) {
    if (!selectedGenerationId) {
      return;
    }

    setIsSubmittingFeedback(true);
    setSuccessMessage("");

    try {
      await feedbackApi.create(selectedGenerationId, payload);
      setSuccessMessage("Отзыв по этой AI-генерации сохранен.");
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setIsSubmittingFeedback(false);
    }
  }

  if (isLoading) {
    return (
      <LoadingState
        title="Загружаем историю генераций"
        description="Получаем запуски, детали выбранной генерации и форму отзыва."
      />
    );
  }

  if (errorMessage && !goal) {
    return (
      <ErrorState
        description={errorMessage}
        actionLabel="Повторить"
        onAction={() => void loadHistory()}
      />
    );
  }

  if (!goal) {
    return (
      <EmptyState
        title="Цель не найдена"
        description="Цель для этой страницы истории недоступна."
      />
    );
  }

  if (generations.length === 0) {
    return (
      <EmptyState
        title="AI-генераций пока нет"
        description="Сгенерируйте первый план в рабочем пространстве цели, чтобы появилась история."
        action={
          <Link to={`/goals/${goal.id}`} className="button button--primary">
            Вернуться к цели
          </Link>
        }
      />
    );
  }

  return (
    <div className="section-stack">
      <section className="card">
        <span className="eyebrow">История цели</span>
        <h2>{goal.title}</h2>
        <p>Сравнивайте AI-запуски, проверяйте ответы модели и оставляйте обратную связь.</p>
      </section>

      {errorMessage ? <ErrorState description={errorMessage} /> : null}

      <div className="history-layout">
        <GenerationHistoryTable
          generations={generations}
          selectedGenerationId={selectedGenerationId}
          onSelectGeneration={handleSelectGeneration}
        />

        <div className="history-panel">
          <section className="card">
            {selectedGeneration ? (
              <>
                <div className="page-section__header">
                  <div>
                    <span className="eyebrow">Выбранная генерация</span>
                    <h3>Детали генерации</h3>
                  </div>
                  <StatusBadge value={selectedGeneration.status} />
                </div>

                <div className="detail-list">
                  <div>
                    <dt>Модель</dt>
                    <dd>{selectedGeneration.modelName}</dd>
                  </div>
                  <div>
                    <dt>Создано</dt>
                    <dd>{formatDateTime(selectedGeneration.createdAt)}</dd>
                  </div>
                </div>

                <div className="section-stack">
                  <div className="metric-card">
                    <span className="metric-label">Промпт</span>
                    <pre>{selectedGeneration.prompt}</pre>
                  </div>
                  <div className="metric-card">
                    <span className="metric-label">Сырой ответ AI</span>
                    <pre>{selectedGeneration.rawAiResponse}</pre>
                  </div>
                  {selectedGeneration.errorMessage ? (
                    <div className="metric-card">
                      <span className="metric-label">Ошибка</span>
                      <p>{selectedGeneration.errorMessage}</p>
                    </div>
                  ) : null}
                </div>
              </>
            ) : (
              <p>Выберите генерацию, чтобы посмотреть детали.</p>
            )}
          </section>

          <section className="card">
            <span className="eyebrow">Обратная связь</span>
            <h3>Оценить генерацию</h3>
            <p>
              Оставьте оценку, чтобы зафиксировать, насколько полезной была эта
              декомпозиция.
            </p>

            {successMessage ? <div className="feedback-success">{successMessage}</div> : null}

            <FeedbackForm
              isSubmitting={isSubmittingFeedback}
              onSubmit={handleSubmitFeedback}
            />
          </section>
        </div>
      </div>
    </div>
  );
}

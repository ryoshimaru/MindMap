import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ApiClientError } from "../api/client";
import {
  analyticsApi,
  decompositionApi,
  goalContextApi,
  goalsApi,
  planVersionsApi,
  tasksApi
} from "../api/endpoints";
import type {
  DecomposeGoalRequest,
  FeasibilityResponse,
  Goal,
  GoalContext,
  GoalContextInput,
  MetricsResponse,
  PlanVersion,
  ProgressResponse,
  QualityResponse,
  TaskStatus,
  TaskTreeNode
} from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { ErrorState } from "../components/ErrorState";
import { FeasibilityCard } from "../components/FeasibilityCard";
import { GoalContextForm } from "../components/GoalContextForm";
import { LoadingState } from "../components/LoadingState";
import { MetricsCard } from "../components/MetricsCard";
import { ProgressCard } from "../components/ProgressCard";
import { QualityScoreCard } from "../components/QualityScoreCard";
import { StatusBadge } from "../components/StatusBadge";
import { TaskTree } from "../components/TaskTree";
import { getErrorMessage } from "../utils/errors";
import { formatDate, formatDateTime } from "../utils/format";

const defaultDecomposePayload: DecomposeGoalRequest = {
  detailLevel: "HIGH",
  includeDependencies: true,
  includeDeadlines: true
};

const emptyContext: GoalContextInput = {
  currentLevel: "",
  constraintsText: "",
  preferences: [],
  expectedResult: ""
};

export function GoalDetailsPage() {
  const { goalId } = useParams<{ goalId: string }>();
  const [goal, setGoal] = useState<Goal | null>(null);
  const [context, setContext] = useState<GoalContext | null>(null);
  const [tasks, setTasks] = useState<TaskTreeNode[]>([]);
  const [progress, setProgress] = useState<ProgressResponse | null>(null);
  const [feasibility, setFeasibility] = useState<FeasibilityResponse | null>(null);
  const [quality, setQuality] = useState<QualityResponse | null>(null);
  const [metrics, setMetrics] = useState<MetricsResponse | null>(null);
  const [versions, setVersions] = useState<PlanVersion[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSavingContext, setIsSavingContext] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [updatingTaskId, setUpdatingTaskId] = useState<string | null>(null);
  const [expandingTaskId, setExpandingTaskId] = useState<string | null>(null);
  const [activatingVersionId, setActivatingVersionId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState("");
  const [actionMessage, setActionMessage] = useState("");

  async function loadGoalWorkspace() {
    if (!goalId) {
      setErrorMessage("В маршруте отсутствует идентификатор цели.");
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setErrorMessage("");

    try {
      const [
        goalResponse,
        contextResponse,
        tasksResponse,
        progressResponse,
        feasibilityResponse,
        qualityResponse,
        metricsResponse,
        versionsResponse
      ] = await Promise.all([
        goalsApi.get(goalId),
        goalContextApi.get(goalId).catch((error: unknown) => {
          if (error instanceof ApiClientError && error.status === 404) {
            return null;
          }

          throw error;
        }),
        tasksApi.listByGoal(goalId),
        analyticsApi.getProgress(goalId),
        analyticsApi.getFeasibility(goalId),
        analyticsApi.getQuality(goalId),
        analyticsApi.getMetrics(goalId),
        planVersionsApi.list(goalId)
      ]);

      setGoal(goalResponse);
      setContext(contextResponse);
      setTasks(tasksResponse);
      setProgress(progressResponse);
      setFeasibility(feasibilityResponse);
      setQuality(qualityResponse);
      setMetrics(metricsResponse);
      setVersions(versionsResponse);
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void loadGoalWorkspace();
  }, [goalId]);

  async function handleSaveContext(payload: GoalContextInput) {
    if (!goalId) {
      return;
    }

    setIsSavingContext(true);
    setActionMessage("");

    try {
      const savedContext = context
        ? await goalContextApi.update(goalId, payload)
        : await goalContextApi.create(goalId, payload);
      setContext(savedContext);
      setActionMessage("Контекст цели сохранен.");
      await loadGoalWorkspace();
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setIsSavingContext(false);
    }
  }

  async function handleRunDecomposition(mode: "initial" | "regenerate") {
    if (!goalId) {
      return;
    }

    setIsGenerating(true);
    setActionMessage("");
    setErrorMessage("");

    try {
      const result =
        mode === "initial"
          ? await decompositionApi.decomposeGoal(goalId, defaultDecomposePayload)
          : await decompositionApi.regenerateGoal(goalId, defaultDecomposePayload);
      setActionMessage(
        `Создано задач: ${result.tasksCreated}. Активная версия плана обновлена.`
      );
      await loadGoalWorkspace();
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setIsGenerating(false);
    }
  }

  async function handleTaskStatusChange(taskId: string, status: TaskStatus) {
    setUpdatingTaskId(taskId);
    setActionMessage("");

    try {
      await tasksApi.updateStatus(taskId, { status });
      setActionMessage("Статус задачи обновлен.");
      await loadGoalWorkspace();
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setUpdatingTaskId(null);
    }
  }

  async function handleDecomposeTask(taskId: string) {
    setExpandingTaskId(taskId);
    setActionMessage("");

    try {
      await decompositionApi.decomposeTask(taskId, defaultDecomposePayload);
      setActionMessage("Задача дополнена AI-подзадачами.");
      await loadGoalWorkspace();
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setExpandingTaskId(null);
    }
  }

  async function handleActivateVersion(versionId: string) {
    if (!goalId) {
      return;
    }

    setActivatingVersionId(versionId);
    setActionMessage("");

    try {
      await planVersionsApi.activate(goalId, versionId);
      setActionMessage("Версия плана активирована.");
      await loadGoalWorkspace();
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setActivatingVersionId(null);
    }
  }

  if (isLoading) {
    return (
      <LoadingState
        title="Загружаем рабочее пространство"
        description="Получаем цель, контекст, аналитику, дерево задач и версии плана."
      />
    );
  }

  if (errorMessage && !goal) {
    return (
      <ErrorState
        description={errorMessage}
        actionLabel="Повторить"
        onAction={() => void loadGoalWorkspace()}
      />
    );
  }

  if (!goal || !progress || !feasibility || !quality || !metrics) {
    return (
      <EmptyState
        title="Данные цели недоступны"
        description="Не удалось загрузить выбранную цель."
      />
    );
  }

  const activeVersion = versions.find((version) => version.isActive) || null;
  const contextValues: GoalContextInput = context
    ? {
        currentLevel: context.currentLevel,
        constraintsText: context.constraintsText,
        preferences: context.preferences,
        expectedResult: context.expectedResult
      }
    : emptyContext;

  return (
    <div className="section-stack">
      <section className="card">
        <div className="page-section__header">
          <div>
            <span className="eyebrow">{goal.category}</span>
            <h2>{goal.title}</h2>
            <p>{goal.description}</p>
          </div>
          <div className="page-actions">
            <StatusBadge value={goal.priority} tone="priority" />
            <StatusBadge value={goal.status} />
          </div>
        </div>

        <div className="detail-list">
          <div>
            <dt>Срок</dt>
            <dd>{formatDate(goal.deadline)}</dd>
          </div>
          <div>
            <dt>Нагрузка в неделю</dt>
            <dd>{goal.availableHoursPerWeek} ч</dd>
          </div>
          <div>
            <dt>Обновлено</dt>
            <dd>{formatDateTime(goal.updatedAt)}</dd>
          </div>
        </div>

        <div className="page-actions">
          {tasks.length === 0 ? (
            <button
              type="button"
              className="button button--primary"
              disabled={isGenerating}
              onClick={() => void handleRunDecomposition("initial")}
            >
              {isGenerating ? "Генерация..." : "Сгенерировать AI-план"}
            </button>
          ) : (
            <button
              type="button"
              className="button button--primary"
              disabled={isGenerating}
              onClick={() => void handleRunDecomposition("regenerate")}
            >
              {isGenerating ? "Перегенерация..." : "Перегенерировать план"}
            </button>
          )}
          <Link to={`/goals/${goal.id}/history`} className="button button--secondary">
            История генераций
          </Link>
        </div>

        {actionMessage ? <div className="feedback-success">{actionMessage}</div> : null}
        {errorMessage ? <ErrorState description={errorMessage} /> : null}
      </section>

      <div className="goal-details-grid">
        <div className="section-stack">
          <section className="card">
            <div className="page-section__header">
              <div>
                <span className="eyebrow">Контекст цели</span>
                <h3>Данные для точной AI-декомпозиции</h3>
              </div>
            </div>
            <GoalContextForm
              initialValues={contextValues}
              isSaving={isSavingContext}
              onSubmit={handleSaveContext}
            />
          </section>

          <section className="card">
            <div className="page-section__header">
              <div>
                <span className="eyebrow">Дерево задач</span>
                <h3>Структура выполнения</h3>
              </div>
            </div>
            <TaskTree
              tasks={tasks}
              updatingTaskId={updatingTaskId}
              expandingTaskId={expandingTaskId}
              onStatusChange={handleTaskStatusChange}
              onDecomposeTask={handleDecomposeTask}
            />
          </section>
        </div>

        <aside className="section-stack">
          <ProgressCard progress={progress} />
          <FeasibilityCard feasibility={feasibility} />
          <QualityScoreCard quality={quality} />
          <MetricsCard metrics={metrics} />

          <section className="card card--compact">
            <div className="page-section__header">
              <div>
                <span className="eyebrow">Версии плана</span>
                <h3>История версий</h3>
              </div>
            </div>

            {activeVersion ? (
              <div className="metric-card">
                <span className="metric-label">Активная версия</span>
                <strong className="metric-value">v{activeVersion.versionNumber}</strong>
                <p>Создано: {formatDateTime(activeVersion.createdAt)}</p>
              </div>
            ) : (
              <p>Активной версии плана пока нет.</p>
            )}

            <div className="section-stack">
              {versions.length === 0 ? (
                <p className="field-hint">
                  После первой AI-декомпозиции версии плана появятся здесь.
                </p>
              ) : (
                versions.map((version) => (
                  <div key={version.id} className="metric-card">
                    <div className="card-header">
                      <strong>Версия {version.versionNumber}</strong>
                      {version.isActive ? <StatusBadge value="ACTIVE" /> : null}
                    </div>
                    <p>Создано: {formatDateTime(version.createdAt)}</p>
                    {!version.isActive ? (
                      <button
                        type="button"
                        className="button button--ghost"
                        disabled={activatingVersionId === version.id}
                        onClick={() => void handleActivateVersion(version.id)}
                      >
                        {activatingVersionId === version.id
                          ? "Активация..."
                          : "Активировать версию"}
                      </button>
                    ) : null}
                  </div>
                ))
              )}
            </div>
          </section>
        </aside>
      </div>
    </div>
  );
}

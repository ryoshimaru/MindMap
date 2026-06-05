import { useEffect, useState, type FormEvent } from "react";
import { aiSettingsApi } from "../api/endpoints";
import type { AIProviderName, AISettingsResponse } from "../api/types";
import { getErrorMessage } from "../utils/errors";
import { LoadingState } from "./LoadingState";

interface AISettingsModalProps {
  onReady?: (settings: AISettingsResponse) => void;
}

export function AISettingsModal({ onReady }: AISettingsModalProps) {
  const [settings, setSettings] = useState<AISettingsResponse | null>(null);
  const [provider, setProvider] = useState<Exclude<AIProviderName, "mock">>("gemini");
  const [apiKey, setApiKey] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  const hasRealKey = Boolean(settings?.maskedKey && settings.provider !== "mock");

  useEffect(() => {
    async function loadSettings() {
      try {
        const currentSettings = await aiSettingsApi.get();
        setSettings(currentSettings);
        if (currentSettings.provider === "gemini" || currentSettings.provider === "deepseek") {
          setProvider(currentSettings.provider);
        }
        if (currentSettings.maskedKey && currentSettings.provider !== "mock") {
          onReady?.(currentSettings);
        }
      } catch (error) {
        setErrorMessage(getErrorMessage(error));
      } finally {
        setIsLoading(false);
      }
    }

    void loadSettings();
  }, [onReady]);

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSaving(true);
    setErrorMessage("");

    try {
      const savedSettings = await aiSettingsApi.save({ provider, apiKey });
      setSettings(savedSettings);
      setApiKey("");
      onReady?.(savedSettings);
    } catch (error) {
      setErrorMessage(getErrorMessage(error));
    } finally {
      setIsSaving(false);
    }
  }

  if (isLoading) {
    return (
      <div className="modal-backdrop">
        <div className="modal-panel">
          <LoadingState
            title="Проверяем AI-провайдера"
            description="Загружаем настройки доступа к AI."
          />
        </div>
      </div>
    );
  }

  if (hasRealKey) return null;

  return (
    <div className="modal-backdrop" role="dialog" aria-modal="true">
      <section className="modal-panel">
        <span className="eyebrow">Настройка AI</span>
        <h2>Подключите AI-провайдера</h2>
        <p>
          Для анализа целей нужен API key Gemini или DeepSeek. Полный ключ не
          отображается после сохранения.
        </p>

        <form className="form-grid" onSubmit={handleSave}>
          <div className="field-group">
            <label className="field-label" htmlFor="ai-provider">
              Провайдер
            </label>
            <select
              id="ai-provider"
              className="field-control"
              value={provider}
              onChange={(event) => setProvider(event.target.value as Exclude<AIProviderName, "mock">)}
            >
              <option value="gemini">Gemini</option>
              <option value="deepseek">DeepSeek</option>
            </select>
          </div>

          <div className="field-group">
            <label className="field-label" htmlFor="ai-api-key">
              API key
            </label>
            <input
              id="ai-api-key"
              className="field-control"
              type="password"
              value={apiKey}
              placeholder="Введите API key"
              required
              onChange={(event) => setApiKey(event.target.value)}
            />
          </div>

          {settings?.maskedKey ? (
            <div className="feedback-success">Сохранённый ключ: {settings.maskedKey}</div>
          ) : null}
          {errorMessage ? <div className="inline-error">{errorMessage}</div> : null}

          <button type="submit" className="button button--primary" disabled={isSaving}>
            {isSaving ? "Сохранение..." : "Сохранить"}
          </button>
        </form>
      </section>
    </div>
  );
}

import { useState } from "react";
import type { AISettingsResponse } from "../api/types";
import { AISettingsModal } from "../components/AISettingsModal";
import { SmartGoalRequestPanel } from "../components/SmartGoalRequestPanel";

export function DashboardPage() {
  const [aiSettings, setAISettings] = useState<AISettingsResponse | null>(null);

  return (
    <div className="dashboard-focus">
      <main className="dashboard-focus__panel" aria-label="Создание цели">
        <SmartGoalRequestPanel />
      </main>
      {!aiSettings?.maskedKey || aiSettings.provider === "mock" ? (
        <AISettingsModal onReady={setAISettings} />
      ) : null}
    </div>
  );
}

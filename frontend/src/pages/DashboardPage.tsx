import { useEffect, useState } from "react";
import { ApiClientError } from "../api/client";
import { aiSettingsApi, profileApi } from "../api/endpoints";
import { AISettingsModal } from "../components/AISettingsModal";
import { ProfileModal } from "../components/ProfileModal";
import { SmartGoalRequestPanel } from "../components/SmartGoalRequestPanel";

export function DashboardPage() {
  const [hasKey, setHasKey] = useState<boolean | null>(null);
  const [hasProfile, setHasProfile] = useState<boolean | null>(null);

  useEffect(() => {
    aiSettingsApi.get().then((settings) => setHasKey(Boolean(settings.maskedKey && settings.provider !== "mock"))).catch(() => setHasKey(false));
    profileApi.get().then(() => setHasProfile(true)).catch((error) => setHasProfile(!(error instanceof ApiClientError && error.status === 404) ? false : false));
  }, []);

  return <div className="create-page">
    <div className="mind-animation" aria-hidden="true"><i/><i/><i/><i/><i/><span/><span/><span/></div>
    <SmartGoalRequestPanel />
    {hasKey === false ? <AISettingsModal onReady={() => setHasKey(true)} /> : null}
    {hasKey && hasProfile === false ? <ProfileModal required onReady={() => setHasProfile(true)} /> : null}
  </div>;
}

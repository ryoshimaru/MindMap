import { humanizeEnum } from "../utils/format";

interface StatusBadgeProps {
  value: string;
  tone?: "status" | "priority" | "feasibility" | "source";
}

const toneMap: Record<string, string> = {
  DRAFT: "muted",
  ACTIVE: "info",
  COMPLETED: "success",
  ARCHIVED: "neutral",
  TODO: "muted",
  IN_PROGRESS: "info",
  DONE: "success",
  CANCELLED: "danger",
  LOW: "neutral",
  MEDIUM: "warning",
  HIGH: "danger",
  REALISTIC: "success",
  RISKY: "warning",
  UNREALISTIC: "danger",
  UNKNOWN: "muted",
  AI: "info",
  MANUAL: "neutral",
  PENDING: "warning",
  SUCCESS: "success",
  FAILED: "danger"
};

export function StatusBadge({ value, tone = "status" }: StatusBadgeProps) {
  const badgeTone = toneMap[value] ?? "neutral";

  return (
    <span className={`status-badge status-badge--${badgeTone}`} data-tone={tone}>
      {humanizeEnum(value)}
    </span>
  );
}

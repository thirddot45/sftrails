import type { TrailCondition } from "@/lib/types";

interface ConditionBadgeProps {
  condition: TrailCondition;
  size?: "sm" | "md" | "lg";
}

const conditionConfig: Record<
  TrailCondition,
  { label: string; className: string; icon: string }
> = {
  dry: {
    label: "Dry",
    className: "bg-green-50 text-green-700 border-green-200",
    icon: "☀",
  },
  wet: {
    label: "Wet",
    className: "bg-blue-50 text-blue-700 border-blue-200",
    icon: "💧",
  },
  muddy: {
    label: "Muddy",
    className: "bg-amber-50 text-amber-700 border-amber-200",
    icon: "🟤",
  },
  snowy: {
    label: "Snowy",
    className: "bg-slate-50 text-slate-700 border-slate-200",
    icon: "❄",
  },
  icy: {
    label: "Icy",
    className: "bg-cyan-50 text-cyan-700 border-cyan-200",
    icon: "🧊",
  },
  unknown: {
    label: "Unknown",
    className: "bg-gray-50 text-gray-600 border-gray-200",
    icon: "?",
  },
};

const sizeClasses = {
  sm: "px-2 py-0.5 text-xs",
  md: "px-2.5 py-1 text-sm",
  lg: "px-3 py-1.5 text-base",
};

export function ConditionBadge({ condition, size = "md" }: ConditionBadgeProps) {
  const config = conditionConfig[condition];

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border font-medium ${config.className} ${sizeClasses[size]}`}
    >
      <span aria-hidden="true">{config.icon}</span>
      <span>{config.label}</span>
    </span>
  );
}

import type { TrailStatus } from "@/lib/types";

interface StatusBadgeProps {
  status: TrailStatus;
  size?: "sm" | "md" | "lg";
}

const statusConfig: Record<
  TrailStatus,
  { label: string; className: string; icon: string }
> = {
  open: {
    label: "Open",
    className: "bg-green-100 text-green-800 border-green-200",
    icon: "✓",
  },
  closed: {
    label: "Closed",
    className: "bg-red-100 text-red-800 border-red-200",
    icon: "✕",
  },
  limited: {
    label: "Limited",
    className: "bg-yellow-100 text-yellow-800 border-yellow-200",
    icon: "⚠",
  },
  unknown: {
    label: "Unknown",
    className: "bg-gray-100 text-gray-800 border-gray-200",
    icon: "?",
  },
};

const sizeClasses = {
  sm: "px-2 py-0.5 text-xs",
  md: "px-2.5 py-1 text-sm",
  lg: "px-3 py-1.5 text-base",
};

export function StatusBadge({ status, size = "md" }: StatusBadgeProps) {
  const config = statusConfig[status];

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border font-medium ${config.className} ${sizeClasses[size]}`}
    >
      <span aria-hidden="true">{config.icon}</span>
      <span>{config.label}</span>
    </span>
  );
}

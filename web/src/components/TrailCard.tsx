import Link from "next/link";
import type { Trail } from "@/lib/types";
import { StatusBadge } from "./StatusBadge";
import { ConditionBadge } from "./ConditionBadge";

interface TrailCardProps {
  trail: Trail;
}

export function TrailCard({ trail }: TrailCardProps) {
  const lastUpdated = new Date(trail.last_updated);
  const timeAgo = getRelativeTime(lastUpdated);

  return (
    <Link
      href={`/trails/${trail.id}`}
      className="block rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-gray-900 truncate">{trail.name}</h3>
          <p className="text-sm text-gray-500 truncate">{trail.park}</p>
        </div>
        <StatusBadge status={trail.status} size="sm" />
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <ConditionBadge condition={trail.condition} size="sm" />
        <span className="text-sm text-gray-500">
          {trail.length_miles} mi • {trail.elevation_gain_ft} ft gain
        </span>
      </div>

      {trail.notes && (
        <p className="mt-2 text-sm text-gray-600 line-clamp-2">{trail.notes}</p>
      )}

      <div className="mt-3 flex items-center justify-between text-xs text-gray-400">
        <span>Updated {timeAgo}</span>
        {trail.is_safe_for_hiking && (
          <span className="text-green-600">Safe for hiking</span>
        )}
      </div>
    </Link>
  );
}

function getRelativeTime(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 60) {
    return `${diffMins}m ago`;
  } else if (diffHours < 24) {
    return `${diffHours}h ago`;
  } else if (diffDays < 7) {
    return `${diffDays}d ago`;
  } else {
    return date.toLocaleDateString();
  }
}

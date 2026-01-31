import type { Trail } from "@/lib/types";
import { TrailCard } from "./TrailCard";

interface TrailListProps {
  trails: Trail[];
  loading?: boolean;
  emptyMessage?: string;
}

export function TrailList({
  trails,
  loading = false,
  emptyMessage = "No trails found",
}: TrailListProps) {
  if (loading) {
    return (
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {[...Array(6)].map((_, i) => (
          <TrailCardSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (trails.length === 0) {
    return (
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-8 text-center">
        <p className="text-gray-500">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {trails.map((trail) => (
        <TrailCard key={trail.id} trail={trail} />
      ))}
    </div>
  );
}

function TrailCardSkeleton() {
  return (
    <div className="animate-pulse rounded-lg border border-gray-200 bg-white p-4">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <div className="h-5 w-3/4 rounded bg-gray-200" />
          <div className="mt-2 h-4 w-1/2 rounded bg-gray-200" />
        </div>
        <div className="h-6 w-16 rounded-full bg-gray-200" />
      </div>
      <div className="mt-3 flex gap-2">
        <div className="h-6 w-16 rounded-full bg-gray-200" />
        <div className="h-6 w-24 rounded bg-gray-200" />
      </div>
      <div className="mt-3 h-4 w-full rounded bg-gray-200" />
    </div>
  );
}

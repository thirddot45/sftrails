import { trailsApi } from "@/lib/api";
import { TrailList } from "@/components";
import type { TrailStatus, TrailCondition } from "@/lib/types";

export const dynamic = "force-dynamic";

interface TrailsPageProps {
  searchParams: Promise<{
    status?: TrailStatus;
    condition?: TrailCondition;
    park?: string;
  }>;
}

export default async function TrailsPage({ searchParams }: TrailsPageProps) {
  const params = await searchParams;
  const { trails, total } = await trailsApi.getTrails({
    status: params.status,
    condition: params.condition,
    park: params.park,
  });

  const activeFilters = Object.entries(params).filter(
    ([, value]) => value !== undefined
  );

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">All Trails</h1>
        <p className="mt-2 text-gray-600">
          {total} trail{total !== 1 ? "s" : ""} found
          {activeFilters.length > 0 && " with active filters"}
        </p>
      </div>

      {/* Filter Pills */}
      {activeFilters.length > 0 && (
        <div className="mb-6 flex flex-wrap gap-2">
          {activeFilters.map(([key, value]) => (
            <span
              key={key}
              className="inline-flex items-center rounded-full bg-green-100 px-3 py-1 text-sm font-medium text-green-800"
            >
              {key}: {value}
            </span>
          ))}
          <a
            href="/trails"
            className="inline-flex items-center rounded-full bg-gray-100 px-3 py-1 text-sm font-medium text-gray-600 hover:bg-gray-200"
          >
            Clear all
          </a>
        </div>
      )}

      {/* Quick Filters */}
      <div className="mb-6 flex flex-wrap gap-2">
        <span className="text-sm font-medium text-gray-500">Quick filters:</span>
        <a
          href="/trails?status=open"
          className="text-sm text-green-600 hover:text-green-700"
        >
          Open
        </a>
        <a
          href="/trails?status=closed"
          className="text-sm text-red-600 hover:text-red-700"
        >
          Closed
        </a>
        <a
          href="/trails?condition=dry"
          className="text-sm text-amber-600 hover:text-amber-700"
        >
          Dry conditions
        </a>
      </div>

      <TrailList trails={trails} emptyMessage="No trails match your filters" />
    </div>
  );
}

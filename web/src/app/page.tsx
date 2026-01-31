import Link from "next/link";
import { trailsApi } from "@/lib/api";
import { StatusSummary, TrailList } from "@/components";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const [summary, trailsResponse] = await Promise.all([
    trailsApi.getStatusSummary(),
    trailsApi.getTrails({ status: "open" }),
  ]);

  const openTrails = trailsResponse.trails.slice(0, 6);

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Hero Section */}
      <section className="mb-12 text-center">
        <h1 className="text-4xl font-bold text-gray-900 sm:text-5xl">
          SF Trail Status
        </h1>
        <p className="mt-4 text-lg text-gray-600">
          Real-time trail conditions for San Francisco area parks
        </p>
      </section>

      {/* Status Summary */}
      <section className="mb-12">
        <h2 className="mb-4 text-xl font-semibold text-gray-900">
          Current Status Overview
        </h2>
        <StatusSummary summary={summary} />
      </section>

      {/* Open Trails */}
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">Open Trails</h2>
          <Link
            href="/trails?status=open"
            className="text-sm font-medium text-green-600 hover:text-green-700"
          >
            View all open trails →
          </Link>
        </div>
        <TrailList trails={openTrails} emptyMessage="No open trails at this time" />
      </section>

      {/* Quick Links */}
      <section className="mt-12 grid gap-4 sm:grid-cols-3">
        <Link
          href="/trails"
          className="rounded-lg border border-gray-200 bg-white p-6 text-center transition-shadow hover:shadow-md"
        >
          <h3 className="font-semibold text-gray-900">All Trails</h3>
          <p className="mt-1 text-sm text-gray-500">
            Browse all {summary.total_trails} trails
          </p>
        </Link>
        <Link
          href="/trails?status=limited"
          className="rounded-lg border border-gray-200 bg-white p-6 text-center transition-shadow hover:shadow-md"
        >
          <h3 className="font-semibold text-gray-900">Limited Access</h3>
          <p className="mt-1 text-sm text-gray-500">
            {summary.limited} trails with restrictions
          </p>
        </Link>
        <Link
          href="/parks"
          className="rounded-lg border border-gray-200 bg-white p-6 text-center transition-shadow hover:shadow-md"
        >
          <h3 className="font-semibold text-gray-900">By Park</h3>
          <p className="mt-1 text-sm text-gray-500">Browse trails by park</p>
        </Link>
      </section>
    </div>
  );
}

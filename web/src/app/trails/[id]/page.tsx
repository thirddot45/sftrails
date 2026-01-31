import Link from "next/link";
import { notFound } from "next/navigation";
import { trailsApi, ApiError } from "@/lib/api";
import { StatusBadge, ConditionBadge } from "@/components";

interface TrailDetailPageProps {
  params: Promise<{ id: string }>;
}

export const dynamic = "force-dynamic";

export default async function TrailDetailPage({ params }: TrailDetailPageProps) {
  const { id } = await params;

  let trail;
  try {
    trail = await trailsApi.getTrail(id);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound();
    }
    throw error;
  }

  const lastUpdated = new Date(trail.last_updated);

  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Breadcrumb */}
      <nav className="mb-6 text-sm">
        <ol className="flex items-center gap-2">
          <li>
            <Link href="/" className="text-gray-500 hover:text-gray-700">
              Home
            </Link>
          </li>
          <li className="text-gray-400">/</li>
          <li>
            <Link href="/trails" className="text-gray-500 hover:text-gray-700">
              Trails
            </Link>
          </li>
          <li className="text-gray-400">/</li>
          <li className="text-gray-900">{trail.name}</li>
        </ol>
      </nav>

      {/* Trail Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{trail.name}</h1>
            <p className="mt-1 text-lg text-gray-600">{trail.park}</p>
          </div>
          <StatusBadge status={trail.status} size="lg" />
        </div>
      </div>

      {/* Status Card */}
      <div className="mb-8 rounded-lg border border-gray-200 bg-white p-6">
        <h2 className="mb-4 text-lg font-semibold text-gray-900">
          Current Conditions
        </h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-gray-500">Trail Condition</p>
            <div className="mt-1">
              <ConditionBadge condition={trail.condition} size="lg" />
            </div>
          </div>
          <div>
            <p className="text-sm text-gray-500">Safe for Hiking</p>
            <p className="mt-1 font-medium">
              {trail.is_safe_for_hiking ? (
                <span className="text-green-600">Yes</span>
              ) : (
                <span className="text-red-600">Use caution</span>
              )}
            </p>
          </div>
        </div>

        {trail.notes && (
          <div className="mt-4 rounded-lg bg-yellow-50 p-4">
            <p className="text-sm font-medium text-yellow-800">Notes</p>
            <p className="mt-1 text-sm text-yellow-700">{trail.notes}</p>
          </div>
        )}
      </div>

      {/* Trail Info */}
      <div className="mb-8 rounded-lg border border-gray-200 bg-white p-6">
        <h2 className="mb-4 text-lg font-semibold text-gray-900">
          Trail Information
        </h2>
        <dl className="grid gap-4 sm:grid-cols-2">
          <div>
            <dt className="text-sm text-gray-500">Length</dt>
            <dd className="mt-1 text-lg font-medium text-gray-900">
              {trail.length_miles} miles
            </dd>
          </div>
          <div>
            <dt className="text-sm text-gray-500">Elevation Gain</dt>
            <dd className="mt-1 text-lg font-medium text-gray-900">
              {trail.elevation_gain_ft} ft
            </dd>
          </div>
          <div>
            <dt className="text-sm text-gray-500">Park</dt>
            <dd className="mt-1">
              <Link
                href={`/parks?name=${encodeURIComponent(trail.park)}`}
                className="text-green-600 hover:text-green-700"
              >
                {trail.park}
              </Link>
            </dd>
          </div>
          <div>
            <dt className="text-sm text-gray-500">Last Updated</dt>
            <dd className="mt-1 text-gray-900">
              {lastUpdated.toLocaleDateString()} at{" "}
              {lastUpdated.toLocaleTimeString()}
            </dd>
          </div>
        </dl>
      </div>

      {/* Back Link */}
      <div className="flex justify-center">
        <Link
          href="/trails"
          className="rounded-lg border border-gray-200 bg-white px-6 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
        >
          ← Back to all trails
        </Link>
      </div>
    </div>
  );
}

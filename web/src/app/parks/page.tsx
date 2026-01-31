import Link from "next/link";
import { trailsApi } from "@/lib/api";

export const dynamic = "force-dynamic";

export default async function ParksPage() {
  const { parks, total } = await trailsApi.getParks();

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Parks</h1>
        <p className="mt-2 text-gray-600">
          Browse trails by park ({total} parks)
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {parks.map((park) => (
          <Link
            key={park.name}
            href={`/trails?park=${encodeURIComponent(park.name)}`}
            className="rounded-lg border border-gray-200 bg-white p-6 transition-shadow hover:shadow-md"
          >
            <h2 className="font-semibold text-gray-900">{park.name}</h2>
            <p className="mt-1 text-sm text-gray-500">
              {park.trail_count} trail{park.trail_count !== 1 ? "s" : ""}
            </p>
          </Link>
        ))}
      </div>
    </div>
  );
}

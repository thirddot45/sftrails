import type { StatusSummary as StatusSummaryType } from "@/lib/types";

interface StatusSummaryProps {
  summary: StatusSummaryType;
}

export function StatusSummary({ summary }: StatusSummaryProps) {
  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
      <StatCard
        label="Open"
        value={summary.open}
        total={summary.total_trails}
        color="green"
      />
      <StatCard
        label="Limited"
        value={summary.limited}
        total={summary.total_trails}
        color="yellow"
      />
      <StatCard
        label="Closed"
        value={summary.closed}
        total={summary.total_trails}
        color="red"
      />
      <StatCard
        label="Total"
        value={summary.total_trails}
        total={summary.total_trails}
        color="gray"
      />
    </div>
  );
}

interface StatCardProps {
  label: string;
  value: number;
  total: number;
  color: "green" | "yellow" | "red" | "gray";
}

const colorClasses = {
  green: "bg-green-50 border-green-200 text-green-700",
  yellow: "bg-yellow-50 border-yellow-200 text-yellow-700",
  red: "bg-red-50 border-red-200 text-red-700",
  gray: "bg-gray-50 border-gray-200 text-gray-700",
};

function StatCard({ label, value, total, color }: StatCardProps) {
  const percentage = total > 0 ? Math.round((value / total) * 100) : 0;

  return (
    <div className={`rounded-lg border p-4 ${colorClasses[color]}`}>
      <p className="text-sm font-medium opacity-80">{label}</p>
      <p className="mt-1 text-2xl font-bold">{value}</p>
      {color !== "gray" && (
        <p className="text-xs opacity-60">{percentage}% of trails</p>
      )}
    </div>
  );
}

/**
 * TypeScript types for the SF Trails API
 */

export type TrailStatus = "open" | "closed" | "limited" | "unknown";
export type TrailCondition = "dry" | "muddy" | "wet" | "snowy" | "icy" | "unknown";

export interface Trail {
  id: string;
  name: string;
  park: string;
  status: TrailStatus;
  condition: TrailCondition;
  length_miles: number;
  elevation_gain_ft: number;
  last_updated: string;
  notes: string;
  is_accessible: boolean;
  is_safe_for_hiking: boolean;
}

export interface TrailListResponse {
  trails: Trail[];
  total: number;
  filters_applied: Record<string, string | null>;
}

export interface StatusSummary {
  total_trails: number;
  open: number;
  closed: number;
  limited: number;
  unknown: number;
  by_condition: Record<string, number>;
}

export interface Park {
  name: string;
  trail_count: number;
}

export interface ParkListResponse {
  parks: Park[];
  total: number;
}

export interface TrailFilters {
  status?: TrailStatus;
  condition?: TrailCondition;
  park?: string;
  max_length_miles?: number;
  max_elevation_gain_ft?: number;
  q?: string;
}

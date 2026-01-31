/**
 * API client for the SF Trails backend
 */

import type {
  Trail,
  TrailListResponse,
  StatusSummary,
  ParkListResponse,
  TrailFilters,
} from "./types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw new ApiError(response.status, `API error: ${response.statusText}`);
  }

  return response.json();
}

function buildQueryString(params: Record<string, string | number | undefined>): string {
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== "") {
      searchParams.append(key, String(value));
    }
  }
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export const trailsApi = {
  /**
   * Get all trails with optional filters
   */
  async getTrails(filters?: TrailFilters): Promise<TrailListResponse> {
    const query = buildQueryString(filters || {});
    return fetchApi<TrailListResponse>(`/api/v1/trails${query}`);
  },

  /**
   * Search trails by name and filters
   */
  async searchTrails(filters?: TrailFilters): Promise<TrailListResponse> {
    const query = buildQueryString(filters || {});
    return fetchApi<TrailListResponse>(`/api/v1/trails/search${query}`);
  },

  /**
   * Get a single trail by ID
   */
  async getTrail(id: string): Promise<Trail> {
    return fetchApi<Trail>(`/api/v1/trails/${encodeURIComponent(id)}`);
  },

  /**
   * Get status summary for all trails
   */
  async getStatusSummary(): Promise<StatusSummary> {
    return fetchApi<StatusSummary>("/api/v1/trails/summary");
  },

  /**
   * Get all parks
   */
  async getParks(): Promise<ParkListResponse> {
    return fetchApi<ParkListResponse>("/api/v1/parks");
  },

  /**
   * Get trails for a specific park
   */
  async getParkTrails(parkName: string): Promise<TrailListResponse> {
    return fetchApi<TrailListResponse>(
      `/api/v1/parks/${encodeURIComponent(parkName)}/trails`
    );
  },
};

export { ApiError };

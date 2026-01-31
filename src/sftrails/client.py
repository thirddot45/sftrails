"""HTTP client for fetching trail data from external sources."""

from typing import Protocol

import httpx

from sftrails.exceptions import DataFetchError


class TrailDataSource(Protocol):
    """Protocol for trail data sources."""

    async def fetch_trails(self) -> list[dict]:
        """Fetch raw trail data from the source."""
        ...

    async def fetch_trail(self, trail_id: str) -> dict | None:
        """Fetch a single trail by ID."""
        ...


class HTTPTrailClient:
    """HTTP client for fetching trail data from a REST API."""

    def __init__(self, base_url: str, timeout: float = 30.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self._client: httpx.AsyncClient | None = None

    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create the HTTP client."""
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(
                base_url=self.base_url,
                timeout=self.timeout,
            )
        return self._client

    async def close(self) -> None:
        """Close the HTTP client."""
        if self._client is not None and not self._client.is_closed:
            await self._client.aclose()
            self._client = None

    async def fetch_trails(self) -> list[dict]:
        """Fetch all trails from the API."""
        client = await self._get_client()
        try:
            response = await client.get("/trails")
            response.raise_for_status()
            return response.json()
        except httpx.HTTPError as e:
            raise DataFetchError(f"Failed to fetch trails: {e}", cause=e)

    async def fetch_trail(self, trail_id: str) -> dict | None:
        """Fetch a single trail by ID."""
        client = await self._get_client()
        try:
            response = await client.get(f"/trails/{trail_id}")
            if response.status_code == 404:
                return None
            response.raise_for_status()
            return response.json()
        except httpx.HTTPError as e:
            raise DataFetchError(f"Failed to fetch trail {trail_id}: {e}", cause=e)

    async def __aenter__(self) -> "HTTPTrailClient":
        """Async context manager entry."""
        await self._get_client()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb) -> None:
        """Async context manager exit."""
        await self.close()


class InMemoryTrailSource:
    """In-memory trail data source for testing and development."""

    def __init__(self, trails: list[dict] | None = None) -> None:
        self._trails: dict[str, dict] = {}
        if trails:
            for trail in trails:
                self._trails[trail["id"]] = trail

    def add_trail(self, trail: dict) -> None:
        """Add a trail to the in-memory store."""
        self._trails[trail["id"]] = trail

    def clear(self) -> None:
        """Clear all trails from the store."""
        self._trails.clear()

    async def fetch_trails(self) -> list[dict]:
        """Fetch all trails from memory."""
        return list(self._trails.values())

    async def fetch_trail(self, trail_id: str) -> dict | None:
        """Fetch a single trail by ID."""
        return self._trails.get(trail_id)

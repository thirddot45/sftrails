"""Trail status service for querying and filtering trails."""

from sftrails.client import TrailDataSource
from sftrails.exceptions import TrailNotFoundError
from sftrails.models import Trail, TrailCondition, TrailStatus


class TrailService:
    """Service for querying trail status information."""

    def __init__(self, data_source: TrailDataSource) -> None:
        self._data_source = data_source
        self._cache: dict[str, Trail] = {}

    async def get_all_trails(self, use_cache: bool = True) -> list[Trail]:
        """Get all trails from the data source."""
        if not use_cache or not self._cache:
            raw_trails = await self._data_source.fetch_trails()
            self._cache = {
                trail["id"]: Trail.from_dict(trail) for trail in raw_trails
            }
        return list(self._cache.values())

    async def get_trail(self, trail_id: str) -> Trail:
        """Get a specific trail by ID."""
        if trail_id in self._cache:
            return self._cache[trail_id]

        raw_trail = await self._data_source.fetch_trail(trail_id)
        if raw_trail is None:
            raise TrailNotFoundError(trail_id)

        trail = Trail.from_dict(raw_trail)
        self._cache[trail_id] = trail
        return trail

    async def get_open_trails(self) -> list[Trail]:
        """Get all trails that are currently open."""
        trails = await self.get_all_trails()
        return [t for t in trails if t.status == TrailStatus.OPEN]

    async def get_accessible_trails(self) -> list[Trail]:
        """Get all trails that are accessible (open or limited)."""
        trails = await self.get_all_trails()
        return [t for t in trails if t.is_accessible()]

    async def get_trails_by_park(self, park: str) -> list[Trail]:
        """Get all trails in a specific park."""
        trails = await self.get_all_trails()
        return [t for t in trails if t.park.lower() == park.lower()]

    async def get_trails_by_condition(
        self, condition: TrailCondition
    ) -> list[Trail]:
        """Get all trails with a specific condition."""
        trails = await self.get_all_trails()
        return [t for t in trails if t.condition == condition]

    async def get_safe_hiking_trails(self) -> list[Trail]:
        """Get all trails safe for hiking."""
        trails = await self.get_all_trails()
        return [t for t in trails if t.is_safe_for_hiking()]

    async def search_trails(
        self,
        status: TrailStatus | None = None,
        condition: TrailCondition | None = None,
        park: str | None = None,
        max_length_miles: float | None = None,
        max_elevation_gain_ft: int | None = None,
    ) -> list[Trail]:
        """Search trails with multiple filter criteria."""
        trails = await self.get_all_trails()

        results = []
        for trail in trails:
            if status is not None and trail.status != status:
                continue
            if condition is not None and trail.condition != condition:
                continue
            if park is not None and trail.park.lower() != park.lower():
                continue
            if max_length_miles is not None and trail.length_miles > max_length_miles:
                continue
            if (
                max_elevation_gain_ft is not None
                and trail.elevation_gain_ft > max_elevation_gain_ft
            ):
                continue
            results.append(trail)

        return results

    def clear_cache(self) -> None:
        """Clear the trail cache."""
        self._cache.clear()

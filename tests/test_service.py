"""Tests for trail service."""

import pytest

from sftrails.client import InMemoryTrailSource
from sftrails.exceptions import TrailNotFoundError
from sftrails.models import TrailCondition, TrailStatus
from sftrails.service import TrailService


class TestTrailService:
    """Tests for TrailService."""

    async def test_get_all_trails(self, trail_service, sample_trail_data):
        """Test getting all trails."""
        trails = await trail_service.get_all_trails()
        assert len(trails) == len(sample_trail_data)

    async def test_get_trail_by_id(self, trail_service):
        """Test getting a specific trail by ID."""
        trail = await trail_service.get_trail("trail-001")
        assert trail.id == "trail-001"
        assert trail.name == "Dipsea Trail"

    async def test_get_nonexistent_trail_raises_error(self, trail_service):
        """Test that getting nonexistent trail raises TrailNotFoundError."""
        with pytest.raises(TrailNotFoundError) as exc_info:
            await trail_service.get_trail("nonexistent")
        assert exc_info.value.trail_id == "nonexistent"

    async def test_get_open_trails(self, trail_service):
        """Test getting only open trails."""
        open_trails = await trail_service.get_open_trails()
        # Based on sample data: trail-001, trail-002, trail-005 are open
        assert len(open_trails) == 3
        for trail in open_trails:
            assert trail.status == TrailStatus.OPEN

    async def test_get_accessible_trails(self, trail_service):
        """Test getting accessible trails (open or limited)."""
        accessible = await trail_service.get_accessible_trails()
        # Based on sample data: 3 open + 1 limited = 4 accessible
        assert len(accessible) == 4
        for trail in accessible:
            assert trail.status in (TrailStatus.OPEN, TrailStatus.LIMITED)

    async def test_get_trails_by_park(self, trail_service):
        """Test filtering trails by park."""
        mt_tam_trails = await trail_service.get_trails_by_park(
            "Mount Tamalpais State Park"
        )
        assert len(mt_tam_trails) == 2
        for trail in mt_tam_trails:
            assert trail.park == "Mount Tamalpais State Park"

    async def test_get_trails_by_park_case_insensitive(self, trail_service):
        """Test that park filtering is case insensitive."""
        trails = await trail_service.get_trails_by_park("mount tamalpais state park")
        assert len(trails) == 2

    async def test_get_trails_by_condition(self, trail_service):
        """Test filtering trails by condition."""
        dry_trails = await trail_service.get_trails_by_condition(TrailCondition.DRY)
        assert len(dry_trails) == 2
        for trail in dry_trails:
            assert trail.condition == TrailCondition.DRY

    async def test_get_safe_hiking_trails(self, trail_service):
        """Test getting trails safe for hiking."""
        safe_trails = await trail_service.get_safe_hiking_trails()
        # Excludes: closed trails and icy/snowy trails
        for trail in safe_trails:
            assert trail.is_safe_for_hiking()

    async def test_search_by_status(self, trail_service):
        """Test searching trails by status."""
        results = await trail_service.search_trails(status=TrailStatus.CLOSED)
        assert len(results) == 1
        assert results[0].id == "trail-003"

    async def test_search_by_condition(self, trail_service):
        """Test searching trails by condition."""
        results = await trail_service.search_trails(condition=TrailCondition.WET)
        assert len(results) == 1
        assert results[0].id == "trail-002"

    async def test_search_by_max_length(self, trail_service):
        """Test searching trails by maximum length."""
        results = await trail_service.search_trails(max_length_miles=3.0)
        for trail in results:
            assert trail.length_miles <= 3.0

    async def test_search_by_max_elevation(self, trail_service):
        """Test searching trails by maximum elevation gain."""
        results = await trail_service.search_trails(max_elevation_gain_ft=500)
        for trail in results:
            assert trail.elevation_gain_ft <= 500

    async def test_search_with_multiple_filters(self, trail_service):
        """Test searching with multiple filter criteria."""
        results = await trail_service.search_trails(
            status=TrailStatus.OPEN,
            condition=TrailCondition.DRY,
            max_length_miles=10.0,
        )
        for trail in results:
            assert trail.status == TrailStatus.OPEN
            assert trail.condition == TrailCondition.DRY
            assert trail.length_miles <= 10.0

    async def test_search_no_matches(self, trail_service):
        """Test search that returns no matches."""
        results = await trail_service.search_trails(
            status=TrailStatus.OPEN,
            condition=TrailCondition.SNOWY,
        )
        assert results == []

    async def test_caching(self, in_memory_source, sample_trail_data):
        """Test that service caches results."""
        service = TrailService(in_memory_source)

        # First call populates cache
        trails1 = await service.get_all_trails()
        assert len(trails1) == len(sample_trail_data)

        # Clear source data
        in_memory_source.clear()

        # Second call should use cache
        trails2 = await service.get_all_trails(use_cache=True)
        assert len(trails2) == len(sample_trail_data)

    async def test_bypass_cache(self, in_memory_source, sample_trail_data):
        """Test bypassing the cache."""
        service = TrailService(in_memory_source)

        # Populate cache
        await service.get_all_trails()

        # Clear source
        in_memory_source.clear()

        # Bypass cache should return empty
        trails = await service.get_all_trails(use_cache=False)
        assert trails == []

    async def test_clear_cache(self, trail_service, in_memory_source):
        """Test clearing the cache."""
        # Populate cache
        await trail_service.get_all_trails()

        # Clear both cache and source
        trail_service.clear_cache()
        in_memory_source.clear()

        # Should now return empty
        trails = await trail_service.get_all_trails()
        assert trails == []

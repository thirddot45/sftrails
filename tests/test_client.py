"""Tests for trail data clients."""

import pytest

from sftrails.client import InMemoryTrailSource


class TestInMemoryTrailSource:
    """Tests for InMemoryTrailSource."""

    @pytest.fixture
    def empty_source(self) -> InMemoryTrailSource:
        """Empty in-memory source."""
        return InMemoryTrailSource()

    @pytest.fixture
    def trail_data(self) -> dict:
        """Single trail data for testing."""
        return {
            "id": "test-001",
            "name": "Test Trail",
            "park": "Test Park",
            "status": "open",
            "condition": "dry",
            "length_miles": 2.0,
            "elevation_gain_ft": 400,
            "last_updated": "2025-01-15T10:00:00",
        }

    async def test_empty_source_returns_empty_list(self, empty_source):
        """Test that empty source returns empty list."""
        trails = await empty_source.fetch_trails()
        assert trails == []

    async def test_add_trail(self, empty_source, trail_data):
        """Test adding a trail to the source."""
        empty_source.add_trail(trail_data)
        trails = await empty_source.fetch_trails()
        assert len(trails) == 1
        assert trails[0]["id"] == "test-001"

    async def test_fetch_trail_by_id(self, empty_source, trail_data):
        """Test fetching a specific trail by ID."""
        empty_source.add_trail(trail_data)
        trail = await empty_source.fetch_trail("test-001")
        assert trail is not None
        assert trail["name"] == "Test Trail"

    async def test_fetch_nonexistent_trail(self, empty_source):
        """Test fetching a trail that doesn't exist."""
        trail = await empty_source.fetch_trail("nonexistent")
        assert trail is None

    async def test_clear_trails(self, empty_source, trail_data):
        """Test clearing all trails."""
        empty_source.add_trail(trail_data)
        empty_source.clear()
        trails = await empty_source.fetch_trails()
        assert trails == []

    async def test_initialize_with_trails(self, sample_trail_data):
        """Test initializing source with trail data."""
        source = InMemoryTrailSource(sample_trail_data)
        trails = await source.fetch_trails()
        assert len(trails) == len(sample_trail_data)

    async def test_overwrite_trail_with_same_id(self, empty_source, trail_data):
        """Test that adding a trail with same ID overwrites."""
        empty_source.add_trail(trail_data)
        updated_data = {**trail_data, "name": "Updated Trail"}
        empty_source.add_trail(updated_data)

        trails = await empty_source.fetch_trails()
        assert len(trails) == 1
        assert trails[0]["name"] == "Updated Trail"

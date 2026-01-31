"""Pytest fixtures for SF Trails tests."""

from datetime import datetime

import pytest

from sftrails.client import InMemoryTrailSource
from sftrails.models import Trail, TrailCondition, TrailStatus
from sftrails.service import TrailService


@pytest.fixture
def sample_trail_data() -> list[dict]:
    """Sample trail data for testing."""
    return [
        {
            "id": "trail-001",
            "name": "Dipsea Trail",
            "park": "Mount Tamalpais State Park",
            "status": "open",
            "condition": "dry",
            "length_miles": 7.4,
            "elevation_gain_ft": 2200,
            "last_updated": "2025-01-15T10:30:00",
            "notes": "Popular trail, expect crowds on weekends",
        },
        {
            "id": "trail-002",
            "name": "Coastal Trail",
            "park": "Golden Gate National Recreation Area",
            "status": "open",
            "condition": "wet",
            "length_miles": 4.2,
            "elevation_gain_ft": 800,
            "last_updated": "2025-01-15T09:00:00",
            "notes": "",
        },
        {
            "id": "trail-003",
            "name": "Steep Ravine Trail",
            "park": "Mount Tamalpais State Park",
            "status": "closed",
            "condition": "muddy",
            "length_miles": 2.1,
            "elevation_gain_ft": 1100,
            "last_updated": "2025-01-14T16:45:00",
            "notes": "Closed due to storm damage",
        },
        {
            "id": "trail-004",
            "name": "Lands End Trail",
            "park": "Golden Gate National Recreation Area",
            "status": "limited",
            "condition": "dry",
            "length_miles": 3.4,
            "elevation_gain_ft": 400,
            "last_updated": "2025-01-15T08:00:00",
            "notes": "Section near Mile Rock closed for maintenance",
        },
        {
            "id": "trail-005",
            "name": "Mount Davidson Trail",
            "park": "Mount Davidson Park",
            "status": "open",
            "condition": "icy",
            "length_miles": 1.2,
            "elevation_gain_ft": 300,
            "last_updated": "2025-01-15T07:30:00",
            "notes": "Ice on north-facing sections",
        },
    ]


@pytest.fixture
def sample_trails(sample_trail_data: list[dict]) -> list[Trail]:
    """Sample Trail objects for testing."""
    return [Trail.from_dict(data) for data in sample_trail_data]


@pytest.fixture
def in_memory_source(sample_trail_data: list[dict]) -> InMemoryTrailSource:
    """In-memory data source with sample trails."""
    return InMemoryTrailSource(sample_trail_data)


@pytest.fixture
def trail_service(in_memory_source: InMemoryTrailSource) -> TrailService:
    """Trail service with in-memory data source."""
    return TrailService(in_memory_source)


@pytest.fixture
def single_trail() -> Trail:
    """A single trail for unit tests."""
    return Trail(
        id="test-001",
        name="Test Trail",
        park="Test Park",
        status=TrailStatus.OPEN,
        condition=TrailCondition.DRY,
        length_miles=2.5,
        elevation_gain_ft=500,
        last_updated=datetime(2025, 1, 15, 10, 0, 0),
        notes="Test notes",
    )

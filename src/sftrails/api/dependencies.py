"""FastAPI dependency injection for the API."""

from functools import lru_cache

from sftrails.client import InMemoryTrailSource, TrailDataSource
from sftrails.service import TrailService

# Sample data for development - in production, use HTTPTrailClient
_SAMPLE_TRAILS = [
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
    {
        "id": "trail-006",
        "name": "Philosopher's Way",
        "park": "McLaren Park",
        "status": "open",
        "condition": "dry",
        "length_miles": 2.7,
        "elevation_gain_ft": 450,
        "last_updated": "2025-01-15T11:00:00",
        "notes": "",
    },
    {
        "id": "trail-007",
        "name": "Tennessee Valley Trail",
        "park": "Golden Gate National Recreation Area",
        "status": "open",
        "condition": "wet",
        "length_miles": 3.8,
        "elevation_gain_ft": 250,
        "last_updated": "2025-01-15T08:30:00",
        "notes": "Beach access available",
    },
]


@lru_cache
def get_data_source() -> TrailDataSource:
    """Get the trail data source (cached singleton)."""
    return InMemoryTrailSource(_SAMPLE_TRAILS)


def get_trail_service() -> TrailService:
    """Get the trail service with injected data source."""
    return TrailService(get_data_source())

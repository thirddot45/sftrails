"""Tests for trail data models."""

from datetime import datetime

import pytest

from sftrails.models import Trail, TrailCondition, TrailStatus


class TestTrailStatus:
    """Tests for TrailStatus enum."""

    def test_status_values(self):
        """Verify all expected status values exist."""
        assert TrailStatus.OPEN.value == "open"
        assert TrailStatus.CLOSED.value == "closed"
        assert TrailStatus.LIMITED.value == "limited"
        assert TrailStatus.UNKNOWN.value == "unknown"

    def test_status_from_string(self):
        """Test creating status from string value."""
        assert TrailStatus("open") == TrailStatus.OPEN
        assert TrailStatus("closed") == TrailStatus.CLOSED


class TestTrailCondition:
    """Tests for TrailCondition enum."""

    def test_condition_values(self):
        """Verify all expected condition values exist."""
        assert TrailCondition.DRY.value == "dry"
        assert TrailCondition.MUDDY.value == "muddy"
        assert TrailCondition.WET.value == "wet"
        assert TrailCondition.SNOWY.value == "snowy"
        assert TrailCondition.ICY.value == "icy"
        assert TrailCondition.UNKNOWN.value == "unknown"


class TestTrail:
    """Tests for Trail dataclass."""

    def test_trail_creation(self, single_trail: Trail):
        """Test basic trail creation."""
        assert single_trail.id == "test-001"
        assert single_trail.name == "Test Trail"
        assert single_trail.park == "Test Park"
        assert single_trail.status == TrailStatus.OPEN
        assert single_trail.condition == TrailCondition.DRY
        assert single_trail.length_miles == 2.5
        assert single_trail.elevation_gain_ft == 500
        assert single_trail.notes == "Test notes"

    def test_is_accessible_open(self):
        """Test is_accessible returns True for open trails."""
        trail = Trail(
            id="t1",
            name="Open Trail",
            park="Park",
            status=TrailStatus.OPEN,
            condition=TrailCondition.DRY,
            length_miles=1.0,
            elevation_gain_ft=100,
            last_updated=datetime.now(),
        )
        assert trail.is_accessible() is True

    def test_is_accessible_limited(self):
        """Test is_accessible returns True for limited trails."""
        trail = Trail(
            id="t1",
            name="Limited Trail",
            park="Park",
            status=TrailStatus.LIMITED,
            condition=TrailCondition.DRY,
            length_miles=1.0,
            elevation_gain_ft=100,
            last_updated=datetime.now(),
        )
        assert trail.is_accessible() is True

    def test_is_accessible_closed(self):
        """Test is_accessible returns False for closed trails."""
        trail = Trail(
            id="t1",
            name="Closed Trail",
            park="Park",
            status=TrailStatus.CLOSED,
            condition=TrailCondition.DRY,
            length_miles=1.0,
            elevation_gain_ft=100,
            last_updated=datetime.now(),
        )
        assert trail.is_accessible() is False

    def test_is_safe_for_hiking_dry_open(self):
        """Test safe hiking check for dry open trail."""
        trail = Trail(
            id="t1",
            name="Safe Trail",
            park="Park",
            status=TrailStatus.OPEN,
            condition=TrailCondition.DRY,
            length_miles=1.0,
            elevation_gain_ft=100,
            last_updated=datetime.now(),
        )
        assert trail.is_safe_for_hiking() is True

    def test_is_safe_for_hiking_icy(self):
        """Test safe hiking check returns False for icy trails."""
        trail = Trail(
            id="t1",
            name="Icy Trail",
            park="Park",
            status=TrailStatus.OPEN,
            condition=TrailCondition.ICY,
            length_miles=1.0,
            elevation_gain_ft=100,
            last_updated=datetime.now(),
        )
        assert trail.is_safe_for_hiking() is False

    def test_is_safe_for_hiking_snowy(self):
        """Test safe hiking check returns False for snowy trails."""
        trail = Trail(
            id="t1",
            name="Snowy Trail",
            park="Park",
            status=TrailStatus.OPEN,
            condition=TrailCondition.SNOWY,
            length_miles=1.0,
            elevation_gain_ft=100,
            last_updated=datetime.now(),
        )
        assert trail.is_safe_for_hiking() is False

    def test_is_safe_for_hiking_closed(self):
        """Test safe hiking check returns False for closed trails."""
        trail = Trail(
            id="t1",
            name="Closed Trail",
            park="Park",
            status=TrailStatus.CLOSED,
            condition=TrailCondition.DRY,
            length_miles=1.0,
            elevation_gain_ft=100,
            last_updated=datetime.now(),
        )
        assert trail.is_safe_for_hiking() is False

    def test_to_dict(self, single_trail: Trail):
        """Test conversion to dictionary."""
        result = single_trail.to_dict()
        assert result["id"] == "test-001"
        assert result["name"] == "Test Trail"
        assert result["status"] == "open"
        assert result["condition"] == "dry"
        assert result["length_miles"] == 2.5
        assert result["elevation_gain_ft"] == 500

    def test_from_dict(self):
        """Test creation from dictionary."""
        data = {
            "id": "t1",
            "name": "Dict Trail",
            "park": "Dict Park",
            "status": "open",
            "condition": "wet",
            "length_miles": 3.5,
            "elevation_gain_ft": 750,
            "last_updated": "2025-01-15T12:00:00",
            "notes": "From dict",
        }
        trail = Trail.from_dict(data)
        assert trail.id == "t1"
        assert trail.name == "Dict Trail"
        assert trail.status == TrailStatus.OPEN
        assert trail.condition == TrailCondition.WET
        assert trail.notes == "From dict"

    def test_from_dict_without_notes(self):
        """Test creation from dictionary without notes field."""
        data = {
            "id": "t1",
            "name": "No Notes Trail",
            "park": "Park",
            "status": "open",
            "condition": "dry",
            "length_miles": 1.0,
            "elevation_gain_ft": 100,
            "last_updated": "2025-01-15T12:00:00",
        }
        trail = Trail.from_dict(data)
        assert trail.notes == ""

    def test_roundtrip_dict_conversion(self, single_trail: Trail):
        """Test that to_dict and from_dict are inverses."""
        data = single_trail.to_dict()
        restored = Trail.from_dict(data)
        assert restored.id == single_trail.id
        assert restored.name == single_trail.name
        assert restored.status == single_trail.status
        assert restored.condition == single_trail.condition

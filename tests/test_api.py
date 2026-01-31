"""Tests for the FastAPI API endpoints."""

import pytest
from fastapi.testclient import TestClient

from sftrails.api.main import app


@pytest.fixture
def client():
    """Create a test client for the API."""
    return TestClient(app)


class TestHealthEndpoint:
    """Tests for the health check endpoint."""

    def test_health_check(self, client):
        """Test health check returns healthy status."""
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert "version" in data


class TestRootEndpoint:
    """Tests for the root endpoint."""

    def test_root(self, client):
        """Test root endpoint returns API info."""
        response = client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert data["name"] == "SF Trails API"
        assert "version" in data
        assert "docs" in data


class TestTrailsEndpoint:
    """Tests for the trails endpoints."""

    def test_list_trails(self, client):
        """Test listing all trails."""
        response = client.get("/api/v1/trails")
        assert response.status_code == 200
        data = response.json()
        assert "trails" in data
        assert "total" in data
        assert len(data["trails"]) == data["total"]

    def test_list_trails_filter_by_status(self, client):
        """Test filtering trails by status."""
        response = client.get("/api/v1/trails?status=open")
        assert response.status_code == 200
        data = response.json()
        for trail in data["trails"]:
            assert trail["status"] == "open"

    def test_list_trails_filter_by_condition(self, client):
        """Test filtering trails by condition."""
        response = client.get("/api/v1/trails?condition=dry")
        assert response.status_code == 200
        data = response.json()
        for trail in data["trails"]:
            assert trail["condition"] == "dry"

    def test_list_trails_filter_by_park(self, client):
        """Test filtering trails by park."""
        response = client.get("/api/v1/trails?park=Mount%20Tamalpais%20State%20Park")
        assert response.status_code == 200
        data = response.json()
        for trail in data["trails"]:
            assert trail["park"] == "Mount Tamalpais State Park"

    def test_list_trails_filter_by_max_length(self, client):
        """Test filtering trails by max length."""
        response = client.get("/api/v1/trails?max_length_miles=3.0")
        assert response.status_code == 200
        data = response.json()
        for trail in data["trails"]:
            assert trail["length_miles"] <= 3.0

    def test_list_trails_combined_filters(self, client):
        """Test filtering trails with multiple filters."""
        response = client.get("/api/v1/trails?status=open&condition=dry")
        assert response.status_code == 200
        data = response.json()
        for trail in data["trails"]:
            assert trail["status"] == "open"
            assert trail["condition"] == "dry"

    def test_get_trail_by_id(self, client):
        """Test getting a specific trail by ID."""
        response = client.get("/api/v1/trails/trail-001")
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == "trail-001"
        assert data["name"] == "Dipsea Trail"
        assert "is_accessible" in data
        assert "is_safe_for_hiking" in data

    def test_get_trail_not_found(self, client):
        """Test getting a non-existent trail."""
        response = client.get("/api/v1/trails/nonexistent")
        assert response.status_code == 404

    def test_search_trails(self, client):
        """Test searching trails by name."""
        response = client.get("/api/v1/trails/search?q=dipsea")
        assert response.status_code == 200
        data = response.json()
        assert len(data["trails"]) >= 1
        assert any("Dipsea" in t["name"] for t in data["trails"])

    def test_search_trails_with_filters(self, client):
        """Test searching trails with filters."""
        response = client.get("/api/v1/trails/search?q=trail&status=open")
        assert response.status_code == 200
        data = response.json()
        for trail in data["trails"]:
            assert trail["status"] == "open"

    def test_status_summary(self, client):
        """Test getting status summary."""
        response = client.get("/api/v1/trails/summary")
        assert response.status_code == 200
        data = response.json()
        assert "total_trails" in data
        assert "open" in data
        assert "closed" in data
        assert "limited" in data
        assert "by_condition" in data
        # Verify counts add up
        assert data["total_trails"] == (
            data["open"] + data["closed"] + data["limited"] + data["unknown"]
        )


class TestParksEndpoint:
    """Tests for the parks endpoints."""

    def test_list_parks(self, client):
        """Test listing all parks."""
        response = client.get("/api/v1/parks")
        assert response.status_code == 200
        data = response.json()
        assert "parks" in data
        assert "total" in data
        for park in data["parks"]:
            assert "name" in park
            assert "trail_count" in park
            assert park["trail_count"] > 0

    def test_get_park_trails(self, client):
        """Test getting trails for a specific park."""
        response = client.get("/api/v1/parks/Mount%20Tamalpais%20State%20Park/trails")
        assert response.status_code == 200
        data = response.json()
        assert len(data["trails"]) > 0
        for trail in data["trails"]:
            assert trail["park"] == "Mount Tamalpais State Park"

    def test_get_park_trails_not_found(self, client):
        """Test getting trails for non-existent park."""
        response = client.get("/api/v1/parks/Nonexistent%20Park/trails")
        assert response.status_code == 404


class TestTrailResponseFields:
    """Tests for trail response field values."""

    def test_trail_has_computed_fields(self, client):
        """Test that trail responses include computed fields."""
        response = client.get("/api/v1/trails/trail-001")
        assert response.status_code == 200
        data = response.json()

        # Check computed fields exist and are booleans
        assert isinstance(data["is_accessible"], bool)
        assert isinstance(data["is_safe_for_hiking"], bool)

    def test_open_trail_is_accessible(self, client):
        """Test that open trails are marked as accessible."""
        response = client.get("/api/v1/trails?status=open")
        assert response.status_code == 200
        data = response.json()
        for trail in data["trails"]:
            assert trail["is_accessible"] is True

    def test_closed_trail_not_accessible(self, client):
        """Test that closed trails are not accessible."""
        response = client.get("/api/v1/trails?status=closed")
        assert response.status_code == 200
        data = response.json()
        for trail in data["trails"]:
            assert trail["is_accessible"] is False

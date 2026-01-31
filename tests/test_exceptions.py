"""Tests for custom exceptions."""

import pytest

from sftrails.exceptions import DataFetchError, SFTrailsError, TrailNotFoundError


class TestSFTrailsError:
    """Tests for base exception."""

    def test_base_error(self):
        """Test base exception can be raised."""
        with pytest.raises(SFTrailsError):
            raise SFTrailsError("Test error")


class TestTrailNotFoundError:
    """Tests for TrailNotFoundError."""

    def test_error_message(self):
        """Test error message includes trail ID."""
        error = TrailNotFoundError("trail-123")
        assert "trail-123" in str(error)

    def test_trail_id_attribute(self):
        """Test trail_id attribute is set."""
        error = TrailNotFoundError("trail-456")
        assert error.trail_id == "trail-456"

    def test_is_sftrails_error(self):
        """Test it's a subclass of SFTrailsError."""
        error = TrailNotFoundError("trail-789")
        assert isinstance(error, SFTrailsError)


class TestDataFetchError:
    """Tests for DataFetchError."""

    def test_error_message(self):
        """Test error message is set."""
        error = DataFetchError("Connection failed")
        assert str(error) == "Connection failed"

    def test_cause_attribute(self):
        """Test cause attribute stores original exception."""
        original = ValueError("Original error")
        error = DataFetchError("Wrapped error", cause=original)
        assert error.cause is original

    def test_cause_none_by_default(self):
        """Test cause is None when not provided."""
        error = DataFetchError("Error without cause")
        assert error.cause is None

    def test_is_sftrails_error(self):
        """Test it's a subclass of SFTrailsError."""
        error = DataFetchError("Fetch failed")
        assert isinstance(error, SFTrailsError)

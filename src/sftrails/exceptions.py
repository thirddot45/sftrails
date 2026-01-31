"""Custom exceptions for SF Trails."""


class SFTrailsError(Exception):
    """Base exception for SF Trails."""

    pass


class TrailNotFoundError(SFTrailsError):
    """Raised when a trail is not found."""

    def __init__(self, trail_id: str) -> None:
        self.trail_id = trail_id
        super().__init__(f"Trail not found: {trail_id}")


class DataFetchError(SFTrailsError):
    """Raised when trail data cannot be fetched."""

    def __init__(self, message: str, cause: Exception | None = None) -> None:
        self.cause = cause
        super().__init__(message)

"""Data models for trail information."""

from dataclasses import dataclass
from datetime import datetime
from enum import Enum


class TrailStatus(Enum):
    """Current operational status of a trail."""

    OPEN = "open"
    CLOSED = "closed"
    LIMITED = "limited"  # Partially open with restrictions
    UNKNOWN = "unknown"


class TrailCondition(Enum):
    """Current condition of the trail surface."""

    DRY = "dry"
    MUDDY = "muddy"
    WET = "wet"
    SNOWY = "snowy"
    ICY = "icy"
    UNKNOWN = "unknown"


@dataclass
class Trail:
    """Represents a hiking/biking trail."""

    id: str
    name: str
    park: str
    status: TrailStatus
    condition: TrailCondition
    length_miles: float
    elevation_gain_ft: int
    last_updated: datetime
    notes: str = ""

    def is_accessible(self) -> bool:
        """Check if the trail is currently accessible for use."""
        return self.status in (TrailStatus.OPEN, TrailStatus.LIMITED)

    def is_safe_for_hiking(self) -> bool:
        """Check if conditions are generally safe for hiking."""
        unsafe_conditions = (TrailCondition.ICY, TrailCondition.SNOWY)
        return self.is_accessible() and self.condition not in unsafe_conditions

    def to_dict(self) -> dict:
        """Convert trail to dictionary representation."""
        return {
            "id": self.id,
            "name": self.name,
            "park": self.park,
            "status": self.status.value,
            "condition": self.condition.value,
            "length_miles": self.length_miles,
            "elevation_gain_ft": self.elevation_gain_ft,
            "last_updated": self.last_updated.isoformat(),
            "notes": self.notes,
        }

    @classmethod
    def from_dict(cls, data: dict) -> "Trail":
        """Create a Trail instance from a dictionary."""
        return cls(
            id=data["id"],
            name=data["name"],
            park=data["park"],
            status=TrailStatus(data["status"]),
            condition=TrailCondition(data["condition"]),
            length_miles=float(data["length_miles"]),
            elevation_gain_ft=int(data["elevation_gain_ft"]),
            last_updated=datetime.fromisoformat(data["last_updated"]),
            notes=data.get("notes", ""),
        )

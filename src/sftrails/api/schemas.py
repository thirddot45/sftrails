"""Pydantic schemas for API request/response models."""

from datetime import datetime
from enum import Enum

from pydantic import BaseModel, ConfigDict, Field


class TrailStatusEnum(str, Enum):
    """Trail status for API responses."""

    OPEN = "open"
    CLOSED = "closed"
    LIMITED = "limited"
    UNKNOWN = "unknown"


class TrailConditionEnum(str, Enum):
    """Trail condition for API responses."""

    DRY = "dry"
    MUDDY = "muddy"
    WET = "wet"
    SNOWY = "snowy"
    ICY = "icy"
    UNKNOWN = "unknown"


class TrailResponse(BaseModel):
    """Response schema for a single trail."""

    id: str
    name: str
    park: str
    status: TrailStatusEnum
    condition: TrailConditionEnum
    length_miles: float = Field(ge=0)
    elevation_gain_ft: int = Field(ge=0)
    last_updated: datetime
    notes: str = ""
    is_accessible: bool
    is_safe_for_hiking: bool

    model_config = ConfigDict(from_attributes=True)


class TrailListResponse(BaseModel):
    """Response schema for a list of trails."""

    trails: list[TrailResponse]
    total: int
    filters_applied: dict[str, str | None] = {}


class StatusSummaryResponse(BaseModel):
    """Response schema for status summary."""

    total_trails: int
    open: int
    closed: int
    limited: int
    unknown: int
    by_condition: dict[str, int]


class ParkResponse(BaseModel):
    """Response schema for a park."""

    name: str
    trail_count: int


class ParkListResponse(BaseModel):
    """Response schema for list of parks."""

    parks: list[ParkResponse]
    total: int


class HealthResponse(BaseModel):
    """Response schema for health check."""

    status: str
    version: str

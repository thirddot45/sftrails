"""Trail API routes."""

from fastapi import APIRouter, Depends, HTTPException, Query

from sftrails.api.dependencies import get_trail_service
from sftrails.api.schemas import (
    ParkListResponse,
    ParkResponse,
    StatusSummaryResponse,
    TrailConditionEnum,
    TrailListResponse,
    TrailResponse,
    TrailStatusEnum,
)
from sftrails.exceptions import TrailNotFoundError
from sftrails.models import Trail, TrailCondition, TrailStatus
from sftrails.service import TrailService

router = APIRouter(prefix="/api/v1/trails", tags=["trails"])


def trail_to_response(trail: Trail) -> TrailResponse:
    """Convert a Trail model to a TrailResponse schema."""
    return TrailResponse(
        id=trail.id,
        name=trail.name,
        park=trail.park,
        status=TrailStatusEnum(trail.status.value),
        condition=TrailConditionEnum(trail.condition.value),
        length_miles=trail.length_miles,
        elevation_gain_ft=trail.elevation_gain_ft,
        last_updated=trail.last_updated,
        notes=trail.notes,
        is_accessible=trail.is_accessible(),
        is_safe_for_hiking=trail.is_safe_for_hiking(),
    )


@router.get("", response_model=TrailListResponse)
async def list_trails(
    status: TrailStatusEnum | None = Query(None, description="Filter by status"),
    condition: TrailConditionEnum | None = Query(None, description="Filter by condition"),
    park: str | None = Query(None, description="Filter by park name"),
    max_length_miles: float | None = Query(None, ge=0, description="Max trail length"),
    max_elevation_gain_ft: int | None = Query(None, ge=0, description="Max elevation gain"),
    service: TrailService = Depends(get_trail_service),
) -> TrailListResponse:
    """List all trails with optional filters."""
    # Convert enum values to model enums if provided
    model_status = TrailStatus(status.value) if status else None
    model_condition = TrailCondition(condition.value) if condition else None

    trails = await service.search_trails(
        status=model_status,
        condition=model_condition,
        park=park,
        max_length_miles=max_length_miles,
        max_elevation_gain_ft=max_elevation_gain_ft,
    )

    return TrailListResponse(
        trails=[trail_to_response(t) for t in trails],
        total=len(trails),
        filters_applied={
            "status": status.value if status else None,
            "condition": condition.value if condition else None,
            "park": park,
            "max_length_miles": str(max_length_miles) if max_length_miles else None,
            "max_elevation_gain_ft": str(max_elevation_gain_ft) if max_elevation_gain_ft else None,
        },
    )


@router.get("/search", response_model=TrailListResponse)
async def search_trails(
    q: str | None = Query(None, description="Search query for trail name"),
    status: TrailStatusEnum | None = Query(None, description="Filter by status"),
    condition: TrailConditionEnum | None = Query(None, description="Filter by condition"),
    park: str | None = Query(None, description="Filter by park name"),
    max_length_miles: float | None = Query(None, ge=0, description="Max trail length"),
    max_elevation_gain_ft: int | None = Query(None, ge=0, description="Max elevation gain"),
    service: TrailService = Depends(get_trail_service),
) -> TrailListResponse:
    """Search trails with query string and filters."""
    model_status = TrailStatus(status.value) if status else None
    model_condition = TrailCondition(condition.value) if condition else None

    trails = await service.search_trails(
        status=model_status,
        condition=model_condition,
        park=park,
        max_length_miles=max_length_miles,
        max_elevation_gain_ft=max_elevation_gain_ft,
    )

    # Apply name search filter if provided
    if q:
        q_lower = q.lower()
        trails = [t for t in trails if q_lower in t.name.lower()]

    return TrailListResponse(
        trails=[trail_to_response(t) for t in trails],
        total=len(trails),
        filters_applied={
            "q": q,
            "status": status.value if status else None,
            "condition": condition.value if condition else None,
            "park": park,
        },
    )


@router.get("/summary", response_model=StatusSummaryResponse)
async def get_status_summary(
    service: TrailService = Depends(get_trail_service),
) -> StatusSummaryResponse:
    """Get aggregate status summary for all trails."""
    trails = await service.get_all_trails()

    status_counts = {"open": 0, "closed": 0, "limited": 0, "unknown": 0}
    condition_counts: dict[str, int] = {}

    for trail in trails:
        status_counts[trail.status.value] += 1
        cond = trail.condition.value
        condition_counts[cond] = condition_counts.get(cond, 0) + 1

    return StatusSummaryResponse(
        total_trails=len(trails),
        open=status_counts["open"],
        closed=status_counts["closed"],
        limited=status_counts["limited"],
        unknown=status_counts["unknown"],
        by_condition=condition_counts,
    )


@router.get("/{trail_id}", response_model=TrailResponse)
async def get_trail(
    trail_id: str,
    service: TrailService = Depends(get_trail_service),
) -> TrailResponse:
    """Get a specific trail by ID."""
    try:
        trail = await service.get_trail(trail_id)
        return trail_to_response(trail)
    except TrailNotFoundError:
        raise HTTPException(status_code=404, detail=f"Trail not found: {trail_id}")


# Parks router
parks_router = APIRouter(prefix="/api/v1/parks", tags=["parks"])


@parks_router.get("", response_model=ParkListResponse)
async def list_parks(
    service: TrailService = Depends(get_trail_service),
) -> ParkListResponse:
    """List all unique parks with trail counts."""
    trails = await service.get_all_trails()

    park_counts: dict[str, int] = {}
    for trail in trails:
        park_counts[trail.park] = park_counts.get(trail.park, 0) + 1

    parks = [
        ParkResponse(name=name, trail_count=count)
        for name, count in sorted(park_counts.items())
    ]

    return ParkListResponse(parks=parks, total=len(parks))


@parks_router.get("/{park_name}/trails", response_model=TrailListResponse)
async def get_park_trails(
    park_name: str,
    service: TrailService = Depends(get_trail_service),
) -> TrailListResponse:
    """Get all trails in a specific park."""
    trails = await service.get_trails_by_park(park_name)

    if not trails:
        # Check if park exists at all
        all_trails = await service.get_all_trails()
        parks = {t.park.lower() for t in all_trails}
        if park_name.lower() not in parks:
            raise HTTPException(status_code=404, detail=f"Park not found: {park_name}")

    return TrailListResponse(
        trails=[trail_to_response(t) for t in trails],
        total=len(trails),
        filters_applied={"park": park_name},
    )

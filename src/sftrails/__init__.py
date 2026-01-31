"""SF Trails - Check trail status in San Francisco area parks."""

from sftrails.models import Trail, TrailStatus, TrailCondition
from sftrails.service import TrailService

__version__ = "0.1.0"
__all__ = ["Trail", "TrailStatus", "TrailCondition", "TrailService"]

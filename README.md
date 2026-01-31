# sftrails

Check trail status in San Francisco area parks.

## Installation

```bash
pip install -e .
```

For development with testing tools:

```bash
pip install -e ".[dev]"
```

## Usage

```python
from sftrails import TrailService, TrailStatus, TrailCondition
from sftrails.client import InMemoryTrailSource

# Create a data source (use HTTPTrailClient for real API)
source = InMemoryTrailSource()

# Create the service
service = TrailService(source)

# Get all open trails
open_trails = await service.get_open_trails()

# Search with filters
results = await service.search_trails(
    status=TrailStatus.OPEN,
    condition=TrailCondition.DRY,
    max_length_miles=5.0,
)
```

## Running Tests

```bash
pytest
```

With coverage:

```bash
pytest --cov=src/sftrails --cov-report=html
```

## Project Structure

```
sftrails/
├── src/sftrails/
│   ├── __init__.py      # Package exports
│   ├── models.py        # Trail, TrailStatus, TrailCondition
│   ├── service.py       # TrailService for querying trails
│   ├── client.py        # Data source clients (HTTP, in-memory)
│   └── exceptions.py    # Custom exceptions
├── tests/
│   ├── conftest.py      # Pytest fixtures
│   ├── test_models.py   # Model tests
│   ├── test_service.py  # Service tests
│   ├── test_client.py   # Client tests
│   └── test_exceptions.py
└── pyproject.toml       # Project configuration
```

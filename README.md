# sftrails

Check trail status in San Francisco area parks.

## Quick Start

### Backend (FastAPI)

```bash
# Install dependencies
pip install -e ".[dev]"

# Run the API server
uvicorn sftrails.api.main:app --reload

# API available at http://localhost:8000
# Docs at http://localhost:8000/docs
```

### Frontend (Next.js)

```bash
cd web

# Install dependencies
npm install

# Create environment file
cp .env.local.example .env.local

# Run the development server
npm run dev

# Frontend available at http://localhost:3000
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/trails` | List all trails with optional filters |
| `GET /api/v1/trails/{id}` | Get a specific trail |
| `GET /api/v1/trails/search` | Search trails by name and filters |
| `GET /api/v1/trails/summary` | Get status summary |
| `GET /api/v1/parks` | List all parks |
| `GET /api/v1/parks/{name}/trails` | Get trails for a park |
| `GET /health` | Health check |

## Running Tests

```bash
# Run all tests
python -m pytest

# Run with coverage
python -m pytest --cov=src/sftrails --cov-report=html
```

## Project Structure

```
sftrails/
├── src/sftrails/
│   ├── __init__.py       # Package exports
│   ├── models.py         # Trail, TrailStatus, TrailCondition
│   ├── service.py        # TrailService for querying trails
│   ├── client.py         # Data source clients (HTTP, in-memory)
│   ├── exceptions.py     # Custom exceptions
│   └── api/
│       ├── main.py       # FastAPI application
│       ├── schemas.py    # Pydantic schemas
│       ├── dependencies.py
│       └── routes/
│           ├── trails.py # Trail endpoints
│           └── health.py # Health check
├── tests/                # Python tests
├── web/                  # Next.js frontend
│   ├── src/
│   │   ├── app/          # Next.js pages
│   │   ├── components/   # React components
│   │   └── lib/          # API client, types
│   └── package.json
└── pyproject.toml
```

## Tech Stack

- **Backend**: Python, FastAPI, Pydantic
- **Frontend**: Next.js 14+, TypeScript, Tailwind CSS
- **Testing**: pytest, pytest-asyncio

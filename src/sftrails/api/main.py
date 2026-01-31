"""FastAPI application for SF Trails API."""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from sftrails.api.routes.health import router as health_router
from sftrails.api.routes.trails import parks_router, router as trails_router

app = FastAPI(
    title="SF Trails API",
    description="API for checking trail status in San Francisco area parks",
    version="0.1.0",
    docs_url="/docs",
    redoc_url="/redoc",
)

# Configure CORS for frontend access
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:3000",  # Next.js dev server
        "http://127.0.0.1:3000",
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(health_router)
app.include_router(trails_router)
app.include_router(parks_router)


@app.get("/")
async def root():
    """Root endpoint with API information."""
    return {
        "name": "SF Trails API",
        "version": "0.1.0",
        "docs": "/docs",
        "health": "/health",
    }

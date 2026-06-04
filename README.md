# SF Trails

Community-driven South Florida mountain bike trail status voting.

## Quick Start

```bash
# Build and run
make run

# Visit http://localhost:8080
```

## Development

```bash
# Generate templ templates
make generate

# Build
make build

# Run tests
make test
```

## Metrics

A privacy-friendly traffic dashboard is available at `/metrics`, protected by
HTTP Basic Auth (both username and password are checked). It shows total visits,
unique visitors, today's counts, a 7-day trend, and top pages. No IP addresses
or personal data are stored or displayed — unique visitors are counted via a
salted, one-way hash.

Configure via environment variables:

- `METRICS_USER` — dashboard username (default: `nimda`)
- `METRICS_PASSWORD` — dashboard password (default: built-in)
- `METRICS_SALT` — salt for the visitor hash (recommended in production)

## Tech Stack

- **Go** with `net/http` standard library router
- **Templ** for type-safe HTML templates
- **HTMX** for dynamic interactions
- **SQLite** (CGo-free) for persistence
- **Tailwind CSS** for styling

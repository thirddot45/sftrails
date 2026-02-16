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

## Tech Stack

- **Go** with `net/http` standard library router
- **Templ** for type-safe HTML templates
- **HTMX** for dynamic interactions
- **SQLite** (CGo-free) for persistence
- **Tailwind CSS** for styling

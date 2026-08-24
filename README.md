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

A privacy-friendly traffic dashboard is available at `/metrics`. It shows total
visits, unique visitors, today's counts, a 7-day breakdown, and top pages. No IP
addresses or personal data are stored or displayed — unique visitors are counted
via a salted, one-way hash.

The page is deliberately bare: aggregate name/value data points only, with no
Tailwind, HTMX, scripts, header, footer, or navigation. It does not use the site
layout, so the only thing a crawler could ever index is the numbers themselves.

The page is open to anyone who visits it, but it is deliberately kept out of
search engines and AI crawlers:

- `robots.txt` disallows `/metrics` in every `User-agent` group, including each
  named AI crawler (a named group replaces the wildcard group entirely, so each
  one needs its own `Disallow`).
- Requests from known search/AI crawler user agents get `403 Forbidden`.
- Responses carry `X-Robots-Tag: noindex, nofollow, noarchive, nosnippet,
  noimageindex, noai, noimageai` and the page emits a matching
  `<meta name="robots">` tag, so anything that fetches it anyway is told not to
  keep it.
- The dashboard publishes no markdown rendition: `/metrics.md` returns `404`,
  and it emits none of the canonical/Open Graph/JSON-LD discovery markup that
  the agent-facing pages use.

Configure via environment variables:

- `METRICS_SALT` — salt for the visitor hash (recommended in production; a random
  per-process salt is used if unset, so unique counts reset on restart)

## Tech Stack

- **Go** with `net/http` standard library router
- **Templ** for type-safe HTML templates
- **HTMX** for dynamic interactions
- **SQLite** (CGo-free) for persistence
- **Tailwind CSS** for styling

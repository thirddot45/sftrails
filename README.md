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

The generated stylesheet is committed, so Go/Docker builds need no Node runtime.
After changing template classes or client-side styles, use Node 22+ to rebuild it:

```bash
npm ci --ignore-scripts
npm run build:css
```

Commit `static/app.min.css` with the source changes. CI runs `npm run check:css`
to catch stale styles. PurgeCSS removes unused selectors from the pinned
Tailwind 2.2.19 source in `assets/`, including all conditional template states.

## Search visibility

Trail names are ordinary HTML links to individual condition pages. Each search
page has a unique title, description, primary heading and canonical URL. The
homepage includes a structured trail list; detail pages include breadcrumbs.
Short park notes link to their park authority, and `/how-it-works` explains the
reporting methodology. Maintain those notes in `templates/trail_guides.go`.

`/sitemap.xml` lists the homepage, methodology and all trail pages. It omits
`lastmod` because no single timestamp captures votes, forecasts and editorial
changes. Markdown alternatives use an HTTP canonical pointing to HTML; both
formats send `Vary: Accept`. API responses, fragments and the health dashboard
send `noindex`. Metrics retains its separate search and AI exclusion policy.

To measure search results, verify `sftrails.info` in Google Search Console and
submit `https://sftrails.info/sitemap.xml`. Inspect the homepage and a trail page
with URL Inspection. Repository changes cannot establish account ownership or
guarantee indexing or ranking. See `docs/seo-review-2026-09-12.md` for findings
and follow-up priorities.

## Metrics

The aggregate traffic dashboard at `/metrics` is public and requires no login.
`METRICS_USER` and `METRICS_PASSWORD` are no longer used and can be removed from
App Platform settings. The page shows totals, daily counts, and top page paths;
it does not expose individual visitor hashes, IP addresses, or user agents.
Snapshots refresh at most once a minute to bound database work.

The dashboard sends `X-Robots-Tag: noindex, nofollow, nosnippet, noarchive` and a
matching HTML meta tag. It has no sitemap, JSON-LD, social, or agent-discovery
entry. `/metrics.md` redirects to the HTML dashboard and `Accept: text/markdown`
does not convert it. Known AI crawler groups in `robots.txt` disallow `/metrics`
(including its suffix/query variants); the response also opts out through
`Content-Signal`. Ordinary search crawlers can fetch the noindex header: blocking
all crawling with robots.txt would prevent them from seeing it. These are
cooperative discovery controls, not access restrictions or a guarantee that a
public URL cannot be fetched, indexed, or used by an uncooperative agent.

Analytics stores an HMAC-SHA256 identifier, keyed by `METRICS_SALT`, and the page
path and timestamp. Set `METRICS_SALT` to a long random **encrypted runtime secret**
in App Platform, shared across replicas. If absent, a random per-process key is
used: restarts/replicas assign different identifiers and may overcount unique
visitors. Rotating the key also changes visitor identifiers; it does not delete
historical counts. Anyone with the key can test candidate IP/User-Agent pairs.
The switch from the previous hash format changes identifiers once on upgrade.

The votes table separately retains IP addresses and client-supplied fingerprints
for one-hour duplicate checks; votes are deleted at UTC midnight while the app
is running. Neither is displayed on the metrics page. Analytics retention is
currently unlimited; these approximate counts can be inflated by automated
traffic and may drop events when the bounded write queue is full.

## Deployment and security

This repository's Docker image targets DigitalOcean App Platform and defaults
`CLIENT_IP_MODE=digitalocean`. It uses the validated, single `DO-Connecting-IP`
header supplied by App Platform ingress, ignoring `X-Forwarded-For` and
`X-Real-IP`. A missing/invalid header falls back to the connection address.
Only use this mode behind App Platform ingress that controls that header.
For local/direct hosting set `CLIENT_IP_MODE=direct` (also the Go binary's
default). Merely sending a header never switches on proxy trust.

`PORT` defaults to 8080. App Platform handles public HTTPS and forwards traffic
to the container's HTTP listener. Responses carry HSTS for this hostname,
anti-framing and MIME-sniffing protections, a referrer policy, and a CSP.
HTMX and Tailwind assets are served locally. The existing inline UI handlers
and styles require `unsafe-inline`; no external scripts or plugins are allowed.

Vote submissions are capped at 4 KiB with fingerprints limited to 64 bytes.
Cross-site browser submissions are rejected using Origin and Fetch Metadata;
non-browser clients remain public. The per-IP limiter permits 30 submissions
per minute, caps tracked addresses, and returns 429 with Retry-After when full.
Limits are per instance; a distributed attacker can still manipulate anonymous
votes. Database transactions serialize each trail's duplicate check and insert
across replicas. Changing the client-supplied fingerprint still changes the
identity used by deduplication; it does not prove a person is unique.

The container runs as a non-root user. Local secrets and database files are
excluded from its build context. CI generates templates, tests both database
backends, runs the race detector and vet, and checks reachable Go vulnerabilities.

## Tech Stack

- **Go** with `net/http` standard library router
- **Templ** for type-safe HTML templates
- **HTMX** for dynamic interactions
- **SQLite** (CGo-free) locally; **PostgreSQL** in the App Platform Docker image
- **Tailwind CSS** for styling

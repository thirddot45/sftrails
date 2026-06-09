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
keyed (HMAC-SHA256) hash of the IP and User-Agent. Because the secret key
(`METRICS_SALT`) is required to recompute a hash, the stored value cannot be
reversed to an IP as long as the key stays secret.

The dashboard ships with **no default credentials** and fails closed: if
`METRICS_USER` or `METRICS_PASSWORD` is unset, `/metrics` returns `503` and is
unreachable. Set both to enable it.

Configure via environment variables:

- `METRICS_USER` — dashboard username (required to enable the dashboard)
- `METRICS_PASSWORD` — dashboard password (required to enable the dashboard)
- `METRICS_SALT` — secret key for the visitor HMAC. Keep it secret and
  unpredictable: anyone who knows it can recompute a visitor's hash from a
  candidate IP. Recommended in production; if unset, a random per-process key is
  used and unique counts reset on restart.

## Deployment & Security

### Trusted proxies

Client IPs gate rate limiting and vote deduplication, so the app must not
blindly trust forwarding headers. `X-Forwarded-For` / `X-Real-IP` are honored
**only** when the direct connection comes from an address you list in
`TRUSTED_PROXIES` (comma-separated CIDRs or IPs, e.g.
`TRUSTED_PROXIES=10.0.0.0/8,172.18.0.1`). When unset, those headers are ignored
and the direct connection IP is used.

- **Behind a reverse proxy / load balancer:** set `TRUSTED_PROXIES` to its
  address(es), or every request will appear to come from the proxy and share a
  single rate-limit/vote bucket.
- **Directly exposed:** leave it unset.

### Security headers

Every response carries a Content-Security-Policy, `X-Content-Type-Options`,
`X-Frame-Options`, `Referrer-Policy`, and `Strict-Transport-Security`. The CSP
allows only same-origin scripts/styles; HTMX and Tailwind are vendored under
`static/vendor/` (rather than loaded from a third-party CDN) so a CDN compromise
can't inject script. The client-side ZIP lookup is allowed to reach
`api.zippopotam.us`.

## Tech Stack

- **Go** with `net/http` standard library router
- **Templ** for type-safe HTML templates
- **HTMX** for dynamic interactions
- **SQLite** (CGo-free) for persistence
- **Tailwind CSS** for styling

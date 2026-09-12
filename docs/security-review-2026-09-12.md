# SF Trails security review — September 12, 2026

Reviewed the application source and deployment files from main at
`f399a6d2016a61c7429cf467aec8362b34eb37a4`, plus the changes in this PR. The existing
`claude/setup-security-engineering-review-63c4rp` branch supplied the local assets,
JSON-LD escaping implementation, and initial request guard/CI work; its proxy
configuration was adapted for DigitalOcean App Platform.

## Findings addressed

| Finding | Change and validation |
| --- | --- |
| Metrics availability did not match the requested public access | Removed Basic Auth and obsolete credential configuration. Full-router tests confirm unauthenticated HTML access, no challenge, and noindex even on errors. The Markdown URL redirects to HTML; Markdown Accept requests do not convert it. |
| Public metrics could be advertised to crawlers | Added X-Robots-Tag and HTML noindex/nofollow/nosnippet/noarchive, disabled discovery/social/JSON-LD metadata for metrics, and added per-AI-crawler robots exclusions. Sitemap and agent documents omit it. Ordinary search crawlers can read noindex. |
| Untrusted/misinterpreted forwarding headers affected voting limits and visitor counts | App Platform mode uses one validated DO-Connecting-IP; direct mode ignores forwarded headers. Tests cover spoofed XFF/X-Real-IP, duplicate/invalid headers, IPv4 normalization, IPv6, and independent client rate limits. |
| Duplicate vote check and insert raced | Per-trail database locking now encloses both operations in one transaction. 24 simultaneous calls over two separate database handles produce one vote in both SQLite and PostgreSQL. |
| Known vulnerabilities in Markdown HTML parsing | Initial govulncheck identified five reachable advisories in golang.org/x/net v0.47.0: GO-2026-5025, 5027, 5028, 5029, and 5030. Updated to v0.56.0, also fixing the unused-package advisory GO-2026-5942. Go builds use 1.26.8. Reachability is a scanner result, not proof these particular handlers were exploitable. |
| JSON-LD string formatting allowed script termination in database-sourced text | Switched to encoding/json escaping. Tests use script-closing markup and validate both emitted JSON-LD documents. Public users cannot currently edit trail descriptions, so exploitation required another path to alter trail data. |
| Missing browser and input safeguards | Added HSTS, CSP, nosniff, anti-framing, referrer policy, 4 KiB vote body limit, 64-byte fingerprint limit, cross-origin browser checks, finite location-coordinate validation, and a header-size bound. Local browser voting works with no logged CSP errors. |
| Public requests could cause unbounded resource use | Rate limiter caps retained addresses and expires idle entries. Analytics caps pending writes at 128. Public metrics serializes snapshot refreshes and reuses results for one minute; cache behavior is tested. |
| Build context and third-party asset exposure | Vendored existing HTMX/Tailwind versions, excluded local secrets/databases from Docker context, and ran the image as non-root. CI action refs are pinned with read-only repository permissions. |
| Privacy claims overstated what was stored | Metrics uses HMAC-SHA256 and exposes only aggregates. README and UI distinguish analytics from raw IP/fingerprint storage in votes and explain key rotation and approximate counts. UTC reset and PostgreSQL date expressions agree. |

## Validation

- Generated templates; Go tests and race detector passed for SQLite and PostgreSQL.
- PostgreSQL integration used a real local PostgreSQL 18.6 server and isolated test schemas.
- Vet and builds passed for both database tags.
- Govulncheck v1.8.0 passed for both tags after dependency updates.
- Gitleaks v8.30.1 found no secrets in the working tree or the 43 commits reachable from fetched branches. No scan can guarantee the absence of every secret.
- Browser inspection verified public metrics rendering and successful HTMX voting against a local test database. No production votes were submitted.
- HTTP checks cover public metrics headers, redirects, discovery documents, and preserved agent SKILL.md routes.

## Remaining limits and deployment considerations

- Metrics is deliberately public. Noindex, robots.txt, and Content-Signal depend on crawler cooperation; they cannot stop arbitrary fetching or guarantee removal of previously indexed copies.
- Anonymous voting is not identity verification. Fingerprints can be changed, IP addresses can be shared, and per-instance rate limits can be distributed across replicas. Stronger abuse controls would require a separate product decision.
- Vote rows still store raw IPs/fingerprints and are cleared by the UTC-midnight scheduler while the app runs. Historical analytics retention remains unlimited. A separate retention/anonymization migration should preserve duplicate-check behavior and existing data deliberately.
- Set METRICS_SALT as an encrypted runtime secret shared across replicas. Without it, per-process keys change visitor identity across restarts/replicas. Its deployment value and existing cloud secrets were not read or changed.
- Trust digitalocean mode only behind App Platform ingress that sets DO-Connecting-IP. The Docker image selects this mode; direct deployments must override it. DigitalOcean account settings, extra proxies, deployment source branch, and database network/TLS settings are outside this source review.
- CSP retains unsafe-inline for existing inline handlers/styles. Frontend dependencies are pinned locally, but this review did not include a separate JavaScript dependency advisory scan or penetration test.

## References

- [DigitalOcean client-IP header](https://docs.digitalocean.com/support/where-can-i-find-the-client-ip-address-of-a-request-connecting-to-my-app/)
- [App Platform HTTPS and HSTS behavior](https://docs.digitalocean.com/products/app-platform/details/limits/)
- [Google noindex and X-Robots-Tag](https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag)
- [Go vulnerability checking](https://go.dev/doc/tutorial/govulncheck)

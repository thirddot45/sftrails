# Security Review: SF Trails Application

**Date:** 2026-03-13
**Application:** SF Trails - Community-driven South Florida mountain bike trail status voting system
**Tech Stack:** Go 1.25, SQLite/PostgreSQL, Templ templates, HTMX, Tailwind CSS
**Review Scope:** Full application codebase

---

## Executive Summary

SF Trails is a low-risk, public-facing web application with no authentication, no user accounts, and no sensitive data handling. The codebase demonstrates solid security fundamentals: parameterized SQL queries, auto-escaping templates, proper input validation, and appropriate server timeouts. The main areas for improvement are the absence of standard security headers and some hardening opportunities around IP trust and vote abuse.

**Overall Risk Level: LOW**

---

## Findings

### CRITICAL: None

### HIGH: None

### MEDIUM

#### M1. Missing Security Headers

**Location:** `main.go:81-88`, `handlers/handlers.go` (all response handlers)
**Description:** The application does not set standard security headers on HTTP responses. While the app is server-rendered and low-risk, these headers provide defense-in-depth against clickjacking, MIME-sniffing, and other browser-level attacks.

**Missing headers:**
- `X-Content-Type-Options: nosniff` - Prevents MIME-type sniffing
- `X-Frame-Options: DENY` - Prevents clickjacking via iframe embedding
- `Content-Security-Policy` - Controls resource loading sources
- `Referrer-Policy: strict-origin-when-cross-origin` - Limits referrer leakage
- `Permissions-Policy` - Restricts browser feature access

**Recommendation:** Add a middleware that sets these headers on all responses:

```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; "+
                "script-src 'self' https://unpkg.com; "+
                "style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
                "img-src 'self'; "+
                "connect-src 'self'; "+
                "frame-ancestors 'none'")
        next.ServeHTTP(w, r)
    })
}
```

---

#### M2. IP Spoofing via X-Forwarded-For Trust

**Location:** `handlers/middleware.go:102-115`
**Description:** The `GetIP()` function blindly trusts `X-Forwarded-For` and `X-Real-IP` headers. Any client can set these headers directly, bypassing the rate limiter and vote deduplication by faking different IP addresses.

```go
func GetIP(r *http.Request) string {
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        parts := strings.Split(xff, ",")
        return strings.TrimSpace(parts[0])  // Trusts client-provided value
    }
    // ...
}
```

**Impact:** An attacker can bypass the 30 req/min rate limit and the 1-hour vote deduplication by sending arbitrary `X-Forwarded-For` headers, enabling vote manipulation.

**Recommendation:**
- If deployed behind a reverse proxy, configure the app to only trust proxy headers from known proxy IPs, or use the rightmost non-private IP from `X-Forwarded-For`.
- If exposed directly to the internet, ignore these headers entirely and use `RemoteAddr`.
- Consider an environment variable (e.g., `TRUSTED_PROXY`) to toggle proxy header trust.

---

### LOW

#### L1. Unbounded Fingerprint Input

**Location:** `handlers/handlers.go:59`, `db/queries.go:143-146`
**Description:** The `fingerprint` form value is accepted and stored with no length validation. While Go and SQLite handle arbitrary-length strings safely, an attacker could submit extremely long fingerprints to waste storage.

```go
fingerprint := r.FormValue("fingerprint")
// Stored directly in database with no length check
```

**Recommendation:** Enforce a maximum length on the fingerprint field (e.g., 64 characters, since the legitimate hash output is 8 hex characters):

```go
if len(fingerprint) > 64 {
    fingerprint = fingerprint[:64]
}
```

---

#### L2. Weak Vote Deduplication

**Location:** `db/queries.go:161-171`, `static/fingerprint.js`
**Description:** Vote deduplication relies on IP + browser fingerprint. The fingerprint is generated client-side from easily spoofed values (user agent, screen size, timezone, language) using a simple djb2 hash. This provides minimal protection against deliberate vote manipulation.

**Impact:** A determined attacker can trivially generate different fingerprints by modifying any component, allowing unlimited votes from the same IP within the 1-hour window.

**Recommendation:** For this application's risk level, the current approach is acceptable. If vote integrity becomes more important, consider:
- Server-side rate limiting per IP regardless of fingerprint (already partially done via rate limiter)
- CAPTCHA for high-frequency voters
- Cookie-based tracking as an additional signal

---

#### L3. In-Memory Rate Limiter Not Persistent

**Location:** `handlers/middleware.go:12-70`
**Description:** The rate limiter stores request counts in an in-memory map. This resets on application restart and doesn't coordinate across multiple instances.

**Impact:** An attacker can reset rate limits by triggering a restart (if they have that ability) or target specific instances in a multi-instance deployment.

**Recommendation:** Acceptable for single-instance deployments. For multi-instance or higher-security needs, use Redis or database-backed rate limiting.

---

#### L4. CDN-Loaded Scripts Without Subresource Integrity (SRI)

**Location:** `templates/layout.templ:52-53`
**Description:** HTMX and Tailwind CSS are loaded from third-party CDNs without SRI hashes:

```html
<script src="https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js" defer></script>
<link href="https://cdn.jsdelivr.net/npm/tailwindcss@2/dist/tailwind.min.css" rel="stylesheet"/>
```

**Impact:** If a CDN is compromised, malicious code could be injected into the page.

**Recommendation:** Add `integrity` and `crossorigin` attributes:
```html
<script src="https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js"
        integrity="sha384-<hash>" crossorigin="anonymous" defer></script>
```
Alternatively, self-host these assets.

---

#### L5. No HTTPS Enforcement in Application

**Location:** `main.go:81-88`
**Description:** The server listens on plain HTTP (`:8080`) with no HSTS header or HTTPS redirect.

**Recommendation:** This is acceptable if HTTPS is terminated at a reverse proxy (common pattern). Ensure the reverse proxy:
- Terminates TLS
- Sets `Strict-Transport-Security` header
- Redirects HTTP to HTTPS

---

### INFORMATIONAL

#### I1. Hardcoded Port

**Location:** `main.go:82`
**Description:** The server port is hardcoded to `:8080`. Consider making it configurable via an environment variable (`PORT`) for deployment flexibility.

---

#### I2. Daily Vote Reset Deletes All Data

**Location:** `main.go:17-35`, `db/queries.go:153-159`
**Description:** The midnight scheduler runs `DELETE FROM votes` which removes all voting history. This is by design but worth noting - there is no vote audit trail.

---

#### I3. Error Messages Don't Leak Internal Details

**Location:** `handlers/handlers.go` (all error paths)
**Description:** (Positive finding) Error responses return generic messages like "Internal server error" while logging detailed errors server-side. This is correct practice.

---

#### I4. SQL Injection Protection

**Location:** `db/queries.go` (all queries)
**Description:** (Positive finding) All queries use parameterized placeholders (`?` for SQLite, `$N` for PostgreSQL). No string interpolation of user input into SQL. The `CHECK` constraint on the vote column provides an additional layer of defense.

---

#### I5. XSS Protection

**Location:** `templates/*.templ`
**Description:** (Positive finding) Templ framework auto-escapes all interpolated values. No use of raw HTML injection patterns.

---

## Positive Security Practices

1. **Parameterized SQL queries** - No SQL injection vectors
2. **Auto-escaping templates** - XSS protection built into framework
3. **Input validation** - Trail IDs must be positive integers, votes must be "open"/"closed"
4. **Rate limiting** - 30 req/min on vote endpoint
5. **Server timeouts** - ReadHeader(5s), Read(10s), Write(15s), Idle(60s) prevent slowloris
6. **Graceful shutdown** - Signal handling with 5s shutdown timeout
7. **Generic error messages** - No internal details leaked to clients
8. **Database constraints** - CHECK, foreign keys, and indexes properly configured
9. **WAL mode** - SQLite configured for concurrent read performance
10. **No eval/exec** - No dynamic code execution patterns
11. **No file uploads** - No file-based attack surface
12. **Structured logging** - slog used consistently with appropriate levels

---

## Recommended Priority Actions

| Priority | Finding | Effort |
|----------|---------|--------|
| 1 | M1: Add security headers middleware | Low |
| 2 | L4: Add SRI to CDN resources | Low |
| 3 | M2: Configure IP trust for proxy deployment | Medium |
| 4 | L1: Add fingerprint length validation | Low |

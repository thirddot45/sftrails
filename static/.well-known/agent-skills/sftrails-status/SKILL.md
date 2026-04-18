---
name: sftrails-status
description: Look up current community-reported status (open/closed) and weather for South Florida mountain bike trails via the SF Trails public JSON API.
---

# SF Trails Status

Use this skill to answer questions like "is Markham Park rideable today?" or
"which South Florida trails are open?" using the SF Trails community voting
API.

## Endpoints

- `GET https://sftrails.info/api/trails` — list every tracked trail with its
  current status, vote counts, location, and weather forecast.
- `GET https://sftrails.info/api/trails/{id}` — fetch a single trail by numeric
  id.

Both return JSON. No authentication is required. A short public cache
(`Cache-Control: public, max-age=60`) is set, so polling faster than once a
minute is wasteful.

## Status values

Each trail's `status` field is one of:

- `open` — ≥3 recent votes and the majority say rideable
- `closed` — ≥3 recent votes and the majority say closed
- `unknown` — insufficient votes in the last 4–12 hours

Votes reset daily at midnight local time, so early-morning `unknown` is
expected.

## Example

```
curl -s https://sftrails.info/api/trails \
  | jq '.trails[] | {name, status, open_votes, closed_votes}'
```

## When NOT to use

- Do not claim a trail is "safe" — this API only reports community votes, not
  official park status. Defer to park authorities for closures, wildlife, or
  emergency conditions.
- Do not cast votes on behalf of users; the `/vote` endpoint is fingerprint
  rate-limited and intended for interactive browser clients.

# SF Trails SEO review — September 12, 2026

## Changes

The live homepage initially contained **zero anchor links** and loaded a
2,934,019-byte stylesheet. Its twelve trail pages were listed in the sitemap
but could not be reached through homepage links. Every page's primary heading
was simply “SF Trails.”

The update adds:

- Server-rendered links to all trail pages, descriptive page headings, unique
  search titles and descriptions, and visible breadcrumbs.
- Short trail notes with park-authority references on eleven detail pages;
  shared routes at Riverbend and the Pinehurst/Okeeheelee naming are explained.
  Tree Tops retains a general note pending more specific route verification.
- A linked methodology page explaining report thresholds, rolling time windows,
  the UTC reset, forecast limitations and the independent community role.
- Structured trail-list and breadcrumb URLs, with JSON escaping retained.
- Absolute sitemap discovery in robots.txt, fourteen canonical sitemap URLs,
  and removal of invented daily modification dates. A database failure returns
  503 instead of publishing an incomplete sitemap.
- Canonical headers for markdown alternatives, consistent `Vary: Accept`, and
  `noindex` on API responses, fragments and the system-health page.
- An **8,332-byte stylesheet**, about **99.7% smaller** before compression,
  generated from the same Tailwind version. Dependencies are locked; CI verifies
  the generated output. This measures stylesheet size, not a Core Web Vitals
  or search-ranking improvement.
- Removal of the trail-card hover movement reported during this work.

Metrics remains accessible without a password and excluded from the sitemap,
structured data and AI discovery. Its noindex headers/meta and AI crawler
rules remain in place.

## Verification

An automated crawl checks every sitemap destination for homepage reachability,
HTTP 200, one primary heading, unique title/description, canonical URL and valid
JSON-LD. Additional checks cover markdown aliases, query variants, genuine
404s, indexing exclusions and sitemap failure behavior. Existing security
tests remain part of the suite.

Local SQLite and PostgreSQL tests passed with the race detector, as did vet,
both builds, Go vulnerability scans and the CSS dependency audit. Desktop and
narrow-screen browser checks cover links, layouts and the public dashboard.

## Next steps requiring account access or ongoing editorial work

1. **Confirm Search Console ownership.** The account/property is unconfirmed.
   Submit the sitemap, inspect the homepage and a representative trail URL,
   then review Page Indexing and Performance. Search queries alone cannot
   establish the site's indexing status.
2. **Validate the park catalogue.** Review seeded map coordinates, city labels,
   designated bike routes and trail names against the park authorities. In
   particular, Virginia Key is within Miami even though the original catalogue
   labels it Key Biscayne; Tree Tops needs more precise bicycle-route detail.
   This update does not migrate the existing trail database or rename URLs.
3. **Build useful first-hand content.** Add dated trail observations, original
   photographs with permission, access notes and verified trailhead information.
   Avoid copying other directories or creating repetitive city landing pages.
4. **Grow real local use.** Accurate recent rider reports and relevant links
   from local clubs are valuable. No outreach or link submissions were sent.
5. **Measure over time.** Compare non-brand impressions, clicks and trail-page
   traffic over several weeks. No ranking or traffic increase has been measured
   or promised, and Search Console submission has not been performed.

## References

- [Google: crawlable links](https://developers.google.com/search/docs/crawling-indexing/links-crawlable)
- [Google: descriptive search titles](https://developers.google.com/search/docs/appearance/title-link)
- [Google: canonical URLs](https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls)
- [Google: sitemap dates and discovery](https://developers.google.com/search/blog/2023/06/sitemaps-lastmod-ping)
- [PurgeCSS: extracting used selectors](https://purgecss.com/extractors)

Park references are linked alongside each note and recorded in
`templates/trail_guides.go`. Fees, opening hours and closure announcements stay
at the source, where the park authority can keep them current.

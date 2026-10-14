# Targeted catalogue verification — 2026-10-13

## Virginia Key

The [City of Miami park listing](https://www.miami.gov/Parks-Public-Places/Parks-Directory/Virginia-Key-Beach-North-Point-Park)
places Virginia Key Beach North Point Park in Miami and describes its mountain
bike network. Correct the original `Key Biscayne` city and location labels to
`Miami`. The existing editorial guide already correctly identifies Miami.

## Tree Tops Park

The [Town of Davie's Pine Island Ridge trail guide](https://www.davie-fl.gov/487/Pine-Island-Ridge-Trail)
describes crossing a paved bike path near Tree Tops' equestrian area and says
maps are available at the entrance gatehouse. Davie's
[park overview](https://www.davie-fl.gov/1493/Things-To-Do-In-Davie)
also lists biking among Tree Tops' activities.

Replace the generic hammock-trail description with designated recreational
cycling information and add a source-linked guide. Do not label its nature or
equestrian trails as mountain-bike singletrack. Exact bicycle-route mileage,
named loops and permission to ride individual unpaved trails remain unverified;
the guide directs riders to park maps and posted bicycle-route signs.

## Applying the corrections

New databases receive corrected seed values. Existing databases receive three
guarded updates during initialization: each changes only its original seeded
field value for the named park. Updates are transactional and idempotent, retain
edited fields, and preserve trail names, IDs, vote associations and URLs.

This verification does not establish accurate trailhead coordinates. Coordinates
remain unchanged rather than substituting a general park map marker for a bike
trailhead. Other catalogue entries still need the broader review identified in
the September SEO report. No production database or deployment was changed while
preparing this patch.

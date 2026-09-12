package templates

import (
	"encoding/json"

	"sftrails/models"
)

// JSON-LD is built with encoding/json rather than string formatting. Marshal
// escapes <, >, and & to \u00XX by default, so the output is safe to embed
// directly inside a <script> block even if trail data ever contains markup
// such as "</script>".

type ldPostalAddress struct {
	Type            string `json:"@type"`
	AddressLocality string `json:"addressLocality"`
	AddressRegion   string `json:"addressRegion"`
	AddressCountry  string `json:"addressCountry"`
}

type ldGeo struct {
	Type      string  `json:"@type"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ldPropertyValue struct {
	Type  string `json:"@type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ldLocation struct {
	Context            string          `json:"@context,omitempty"`
	Type               string          `json:"@type"`
	ID                 string          `json:"@id"`
	URL                string          `json:"url"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Address            ldPostalAddress `json:"address"`
	Geo                ldGeo           `json:"geo"`
	PublicAccess       bool            `json:"publicAccess"`
	AdditionalProperty ldPropertyValue `json:"additionalProperty"`
}

type ldWebsite struct {
	Context     string     `json:"@context"`
	Type        string     `json:"@type"`
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	Description string     `json:"description"`
	MainEntity  ldItemList `json:"mainEntity"`
}

type ldListItem struct {
	Type     string `json:"@type"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	Item     string `json:"item"`
}

type ldItemList struct {
	Context string       `json:"@context,omitempty"`
	Type    string       `json:"@type"`
	Items   []ldListItem `json:"itemListElement"`
}

func statusLabel(s models.TrailStatus) string {
	switch s {
	case models.StatusOpen:
		return "Open"
	case models.StatusClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}

// locationLD builds the SportsActivityLocation node for a trail. withContext
// adds the top-level @context, used when the node stands alone (detail page)
// rather than nested under a parent document (index page).
func locationLD(t models.TrailWithStatus, withContext bool) ldLocation {
	ld := ldLocation{
		Type:        "SportsActivityLocation",
		ID:          "https://sftrails.info/trail/" + slugify(t.Name) + "#place",
		URL:         "https://sftrails.info/trail/" + slugify(t.Name),
		Name:        t.Name,
		Description: t.Description,
		Address: ldPostalAddress{
			Type:            "PostalAddress",
			AddressLocality: t.City,
			AddressRegion:   "FL",
			AddressCountry:  "US",
		},
		Geo: ldGeo{
			Type:      "GeoCoordinates",
			Latitude:  t.Latitude,
			Longitude: t.Longitude,
		},
		PublicAccess: true,
		AdditionalProperty: ldPropertyValue{
			Type:  "PropertyValue",
			Name:  "Trail Status",
			Value: statusLabel(t.Status),
		},
	}
	if withContext {
		ld.Context = "https://schema.org"
	}
	return ld
}

// trailsJSONLDScript returns the full <script> block for the index page.
func trailsJSONLDScript(trails []models.TrailWithStatus) string {
	items := make([]ldListItem, 0, len(trails))
	for i, t := range trails {
		items = append(items, ldListItem{Type: "ListItem", Position: i + 1, Name: t.Name, Item: "https://sftrails.info/trail/" + slugify(t.Name)})
	}
	doc := ldWebsite{
		Context:     "https://schema.org",
		Type:        "WebSite",
		Name:        "SF Trails",
		URL:         "https://sftrails.info",
		Description: "Community-driven South Florida mountain bike trail status reports",
		MainEntity:  ldItemList{Type: "ItemList", Items: items},
	}
	return wrapJSONLD(doc)
}

// trailDetailJSONLD returns the full <script> block for a single trail page.
func trailDetailJSONLD(t models.TrailWithStatus) string {
	return wrapJSONLD(locationLD(t, true))
}

func trailBreadcrumbJSONLD(t models.TrailWithStatus) string {
	return wrapJSONLD(ldItemList{
		Context: "https://schema.org",
		Type:    "BreadcrumbList",
		Items: []ldListItem{
			{Type: "ListItem", Position: 1, Name: "South Florida trails", Item: "https://sftrails.info/"},
			{Type: "ListItem", Position: 2, Name: t.Name, Item: "https://sftrails.info/trail/" + slugify(t.Name)},
		},
	})
}

func wrapJSONLD(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `<script type="application/ld+json">{}</script>`
	}
	return `<script type="application/ld+json">` + string(b) + `</script>`
}

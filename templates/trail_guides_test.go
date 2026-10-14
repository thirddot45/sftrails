package templates

import (
	"strings"
	"testing"
)

func TestTreeTopsGuideUsesVerifiedBicycleAccess(t *testing.T) {
	guide, ok := trailGuides["tree-tops-park"]
	if !ok {
		t.Fatal("Tree Tops guide is missing")
	}
	for _, detail := range []string{"paved bike path", "entrance gatehouse", "posted bicycle-route signs"} {
		if !strings.Contains(guide.Summary, detail) {
			t.Errorf("Tree Tops guide is missing %q", detail)
		}
	}
	if guide.URL != "https://www.davie-fl.gov/487/Pine-Island-Ridge-Trail" {
		t.Errorf("unexpected Tree Tops source: %s", guide.URL)
	}
}

package templates

import (
	"encoding/json"
	"strings"
	"testing"

	"sftrails/models"
)

func TestJSONLDEscapesScriptBreakout(t *testing.T) {
	trail := models.TrailWithStatus{}
	trail.Name = `</script><script>alert(1)</script>`
	trail.Description = `quotes " and <img src=x onerror=alert(1)>`
	for _, output := range []string{trailsJSONLDScript([]models.TrailWithStatus{trail}), trailDetailJSONLD(trail), trailBreadcrumbJSONLD(trail)} {
		if strings.Count(output, "</script>") != 1 || strings.Contains(output, "<img") {
			t.Fatalf("unsafe script output: %s", output)
		}
		payload := strings.TrimSuffix(strings.TrimPrefix(output, `<script type="application/ld+json">`), "</script>")
		if !json.Valid([]byte(payload)) {
			t.Error("invalid JSON-LD")
		}
	}
}

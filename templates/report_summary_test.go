package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"sftrails/models"
)

func TestTrailCardReportSummary(t *testing.T) {
	stamp := time.Date(2026, 10, 14, 8, 30, 0, 0, time.FixedZone("EDT", -4*60*60))
	cases := []struct {
		name         string
		trail        models.TrailWithStatus
		want, absent []string
	}{
		{"empty", models.TrailWithStatus{}, []string{"No recent reports", "past 12h"}, []string{"0/0", "Last report:", ">Unknown<"}},
		{"stale", models.TrailWithStatus{LastReportAt: stamp}, []string{"No recent reports", "Last report:", `datetime="2026-10-14T12:30:00Z"`, "Oct 14, 2026 at 12:30 PM UTC"}, nil},
		{"insufficient", models.TrailWithStatus{Status: models.StatusUnknown, TotalVotes: 1, OpenVotes: 1, ReportWindowHours: 12, LastReportAt: stamp}, []string{"1 rideable · 0 closed", "past 12h", "At least 3 reports needed", ">Unknown<"}, []string{"No recent reports", "Reports are tied"}},
		{"tie", models.TrailWithStatus{Status: models.StatusUnknown, TotalVotes: 4, OpenVotes: 2, ClosedVotes: 2, ReportWindowHours: 4}, []string{"2 rideable · 2 closed", "past 4h", "Reports are tied"}, []string{"At least 3 reports needed"}},
		{"open", models.TrailWithStatus{Status: models.StatusOpen, TotalVotes: 3, OpenVotes: 2, ClosedVotes: 1, ReportWindowHours: 4}, []string{"2 rideable · 1 closed", "past 4h", ">Open<"}, []string{"Reports are tied", "At least 3 reports needed"}},
		{"closed", models.TrailWithStatus{Status: models.StatusClosed, TotalVotes: 3, OpenVotes: 1, ClosedVotes: 2, ReportWindowHours: 12}, []string{"1 rideable · 2 closed", "past 12h", ">Closed<"}, []string{"No recent reports"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.trail.ID = 1
			tc.trail.Name = "Amelia Earhart Park"
			var out bytes.Buffer
			if err := TrailCard(tc.trail).Render(context.Background(), &out); err != nil {
				t.Fatal(err)
			}
			body := out.String()
			for _, want := range append(tc.want, `id="trail-1"`, `href="/trail/amelia-earhart-park"`, `hx-post="/vote"`, `hx-target="#trail-1"`) {
				if !strings.Contains(body, want) {
					t.Errorf("missing %q in %s", want, body)
				}
			}
			for _, absent := range tc.absent {
				if strings.Contains(body, absent) {
					t.Errorf("unexpected %q", absent)
				}
			}
		})
	}
}

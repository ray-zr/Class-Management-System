package logic

import (
	"testing"
	"time"
)

func TestRankingRangeCustomDatesAreInclusive(t *testing.T) {
	start, end, total, err := rankingRange(time.Now(), "", "2026-08-01", "2026-08-25", false)
	if err != nil {
		t.Fatalf("rankingRange() error = %v", err)
	}
	if total {
		t.Fatal("rankingRange() unexpectedly returned total mode")
	}
	if got, want := start.Format("2006-01-02 15:04:05"), "2026-08-01 00:00:00"; got != want {
		t.Fatalf("start = %q, want %q", got, want)
	}
	if got, want := end.Format("2006-01-02 15:04:05"), "2026-08-26 00:00:00"; got != want {
		t.Fatalf("end = %q, want %q", got, want)
	}
	if start.Location() != time.Local || end.Location() != time.Local {
		t.Fatalf("range location = (%v, %v), want time.Local", start.Location(), end.Location())
	}
}

func TestRankingRangeRejectsInvalidRanges(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
	}{
		{name: "only start", startDate: "2026-08-01"},
		{name: "only end", endDate: "2026-08-25"},
		{name: "end before start", startDate: "2026-08-25", endDate: "2026-08-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, _, err := rankingRange(time.Now(), "", tt.startDate, tt.endDate, false); err == nil {
				t.Fatal("rankingRange() accepted an invalid range")
			}
		})
	}
}

func TestRankingRangeTotalIgnoresFilters(t *testing.T) {
	start, end, total, err := rankingRange(time.Now(), "invalid", "invalid", "invalid", true)
	if err != nil {
		t.Fatalf("rankingRange() error = %v", err)
	}
	if !total || !start.IsZero() || !end.IsZero() {
		t.Fatalf("rankingRange() = (%v, %v, %v), want zero range in total mode", start, end, total)
	}
}

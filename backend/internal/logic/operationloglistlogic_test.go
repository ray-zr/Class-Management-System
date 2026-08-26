package logic

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"class-management-system/backend/internal/httperr"
)

func TestOperationLogRangeIsInclusive(t *testing.T) {
	start, end, err := operationLogRange("2026-08-01", "2026-08-26")
	if err != nil {
		t.Fatalf("operationLogRange() error = %v", err)
	}
	if got, want := start.Format("2006-01-02 15:04:05"), "2026-08-01 00:00:00"; got != want {
		t.Fatalf("start = %q, want %q", got, want)
	}
	if got, want := end.Format("2006-01-02 15:04:05"), "2026-08-27 00:00:00"; got != want {
		t.Fatalf("end = %q, want %q", got, want)
	}
	if start.Location() != time.Local || end.Location() != time.Local {
		t.Fatalf("range location = (%v, %v), want time.Local", start.Location(), end.Location())
	}
}

func TestOperationLogRangeAcceptsEmptyRange(t *testing.T) {
	start, end, err := operationLogRange("", "")
	if err != nil || !start.IsZero() || !end.IsZero() {
		t.Fatalf("operationLogRange() = (%v, %v, %v), want zero range", start, end, err)
	}
}

func TestOperationLogRangeRejectsInvalidRanges(t *testing.T) {
	tests := []struct {
		start string
		end   string
	}{
		{start: "2026-08-01"},
		{end: "2026-08-26"},
		{start: "invalid", end: "2026-08-26"},
		{start: "2026-08-27", end: "2026-08-26"},
	}
	for _, tt := range tests {
		_, _, err := operationLogRange(tt.start, tt.end)
		var httpErr *httperr.Error
		if !errors.As(err, &httpErr) || httpErr.Code != http.StatusBadRequest {
			t.Fatalf("operationLogRange(%q, %q) error = %v, want HTTP 400", tt.start, tt.end, err)
		}
	}
}

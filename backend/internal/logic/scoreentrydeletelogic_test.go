package logic

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/types"
)

func TestScoreEntryDeleteRequiresValidReason(t *testing.T) {
	tests := []struct {
		name string
		id   int64
		req  *types.ScoreEntryRevokeReq
	}{
		{name: "invalid id", req: &types.ScoreEntryRevokeReq{Reason: "reason"}},
		{name: "nil request", id: 1},
		{name: "blank reason", id: 1, req: &types.ScoreEntryRevokeReq{Reason: "  "}},
		{name: "reason too long", id: 1, req: &types.ScoreEntryRevokeReq{Reason: strings.Repeat("撤", 256)}},
	}
	logic := &ScoreEntryDeleteLogic{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := logic.ScoreEntryDelete(tt.id, tt.req)
			var httpErr *httperr.Error
			if !errors.As(err, &httpErr) || httpErr.Code != http.StatusBadRequest {
				t.Fatalf("ScoreEntryDelete() error = %v, want HTTP 400", err)
			}
		})
	}
}

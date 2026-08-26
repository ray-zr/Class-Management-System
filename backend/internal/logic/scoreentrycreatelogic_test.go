package logic

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/types"
)

func TestScoreEntryCreateRejectsInvalidRequestsBeforeDatabaseAccess(t *testing.T) {
	tests := []struct {
		name string
		req  *types.ScoreEntryCreateReq
	}{
		{name: "nil request"},
		{name: "missing score item", req: &types.ScoreEntryCreateReq{Scope: "class"}},
		{name: "invalid scope", req: &types.ScoreEntryCreateReq{Scope: "invalid", ScoreItemId: 1}},
		{name: "missing target", req: &types.ScoreEntryCreateReq{Scope: "student", ScoreItemId: 1}},
	}
	logic := &ScoreEntryCreateLogic{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := logic.ScoreEntryCreate(tt.req)
			var httpErr *httperr.Error
			if !errors.As(err, &httpErr) || httpErr.Code != http.StatusBadRequest {
				t.Fatalf("ScoreEntryCreate() error = %v, want HTTP 400", err)
			}
		})
	}
}

func TestScoreRequestFingerprint(t *testing.T) {
	base := &types.ScoreEntryCreateReq{Scope: "student", TargetId: 1, ScoreItemId: 2, Remark: "good"}
	withDifferentRequestID := *base
	withDifferentRequestID.RequestId = "another-id"
	if scoreRequestFingerprint(base) != scoreRequestFingerprint(&withDifferentRequestID) {
		t.Fatal("requestId must not change the operation fingerprint")
	}

	changed := *base
	changed.Remark = "different"
	if scoreRequestFingerprint(base) == scoreRequestFingerprint(&changed) {
		t.Fatal("different scoring payloads must have different fingerprints")
	}
}

func TestScoreEntryRespIncludesHistoricalSnapshots(t *testing.T) {
	createdAt := time.Unix(1_700_000_000, 0)
	revokedAt := createdAt.Add(time.Hour)
	entry := &model.ScoreEntry{
		BaseModel:             model.BaseModel{ID: 7, CreatedAt: createdAt},
		StudentID:             11,
		GroupID:               12,
		DimensionID:           13,
		ScoreItemID:           14,
		Score:                 5,
		Remark:                "remark",
		RequestID:             "request-id",
		StudentNoSnapshot:     "2026001",
		StudentNameSnapshot:   "张三",
		GroupNameSnapshot:     "第一组",
		DimensionNameSnapshot: "课堂表现",
		ScoreItemNameSnapshot: "积极发言",
		RevokedAt:             &revokedAt,
		RevokeReason:          "录入对象错误",
	}
	resp := scoreEntryResp(entry)
	if resp.Id != entry.ID || resp.RequestId != entry.RequestID || resp.CreatedAt != createdAt.Unix() {
		t.Fatalf("scoreEntryResp() base fields = %+v", resp)
	}
	if resp.StudentNameSnapshot != entry.StudentNameSnapshot || resp.GroupNameSnapshot != entry.GroupNameSnapshot ||
		resp.DimensionNameSnapshot != entry.DimensionNameSnapshot || resp.ScoreItemNameSnapshot != entry.ScoreItemNameSnapshot {
		t.Fatalf("scoreEntryResp() snapshots = %+v", resp)
	}
	if resp.RevokedAt != revokedAt.Unix() || resp.RevokeReason != entry.RevokeReason {
		t.Fatalf("scoreEntryResp() revoke fields = %+v", resp)
	}
}

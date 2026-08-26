package logic

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/types"
)

func badRequest(msg string) error {
	return &httperr.Error{Code: http.StatusBadRequest, Msg: msg}
}

func scoreRequestFingerprint(req *types.ScoreEntryCreateReq) string {
	payload := fmt.Sprintf("%s\x00%d\x00%d\x00%s", req.Scope, req.TargetId, req.ScoreItemId, req.Remark)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func scoreEntryResp(entry *model.ScoreEntry) *types.ScoreEntryResp {
	if entry == nil {
		return nil
	}
	resp := &types.ScoreEntryResp{
		Id:                    entry.ID,
		StudentId:             entry.StudentID,
		GroupId:               entry.GroupID,
		DimensionId:           entry.DimensionID,
		ScoreItemId:           entry.ScoreItemID,
		Score:                 entry.Score,
		Remark:                entry.Remark,
		RequestId:             entry.RequestID,
		StudentNoSnapshot:     entry.StudentNoSnapshot,
		StudentNameSnapshot:   entry.StudentNameSnapshot,
		GroupNameSnapshot:     entry.GroupNameSnapshot,
		DimensionNameSnapshot: entry.DimensionNameSnapshot,
		ScoreItemNameSnapshot: entry.ScoreItemNameSnapshot,
		CreatedAt:             entry.CreatedAt.Unix(),
		RevokeReason:          entry.RevokeReason,
	}
	if entry.RevokedAt != nil {
		resp.RevokedAt = entry.RevokedAt.Unix()
	}
	return resp
}

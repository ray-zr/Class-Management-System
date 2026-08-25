// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ScoreEntryCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreEntryCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreEntryCreateLogic {
	return &ScoreEntryCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScoreEntryCreateLogic) ScoreEntryCreate(req *types.ScoreEntryCreateReq) (resp *types.ScoreEntryResp, err error) {
	if req == nil || req.ScoreItemId <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	if req.Scope != "student" && req.Scope != "group" && req.Scope != "class" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid scope"}
	}
	if req.Scope != "class" && req.TargetId <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing targetId"}
	}
	if utf8.RuneCountInString(req.Remark) > 255 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "remark is too long"}
	}
	requestID := strings.TrimSpace(req.RequestId)
	if requestID == "" {
		requestID = uuid.NewString()
	}
	if len(requestID) > 64 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "requestId is too long"}
	}
	requestCopy := *req
	requestCopy.RequestId = requestID
	fingerprint := scoreRequestFingerprint(&requestCopy)

	var out *types.ScoreEntryResp
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		studentRepo := repository.NewStudentRepo(tx)
		groupRepo := repository.NewGroupRepo(tx)
		scoreItemRepo := repository.NewScoreItemRepo(tx)
		scoreEntryRepo := repository.NewScoreEntryRepo(tx)
		scoreOperationRepo := repository.NewScoreOperationRepo(tx)
		recentRepo := repository.NewRecentScoreItemRepo(tx)
		created, operation, err := scoreOperationRepo.CreateOrGet(l.ctx, &model.ScoreOperation{
			RequestID:   requestID,
			Fingerprint: fingerprint,
		})
		if err != nil {
			return err
		}
		if !created {
			if operation.Fingerprint != fingerprint {
				return &httperr.Error{Code: http.StatusConflict, Msg: "requestId was already used for another request"}
			}
			if operation.LastEntryID <= 0 {
				return &httperr.Error{Code: http.StatusConflict, Msg: "request is still being processed"}
			}
			entry, err := scoreEntryRepo.Get(l.ctx, operation.LastEntryID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return &httperr.Error{Code: http.StatusConflict, Msg: "original score entry is no longer available"}
				}
				return err
			}
			out = scoreEntryResp(entry)
			return nil
		}

		item, err := scoreItemRepo.GetForUpdate(l.ctx, req.ScoreItemId)
		if err != nil {
			return err
		}
		dimension, err := repository.NewDimensionRepo(tx).Get(l.ctx, item.DimensionID)
		if err != nil {
			return err
		}

		var students []model.Student
		switch req.Scope {
		case "student":
			student, err := studentRepo.GetForUpdate(l.ctx, req.TargetId)
			if err != nil {
				return err
			}
			students = []model.Student{*student}
		case "group":
			if _, err := groupRepo.GetForUpdate(l.ctx, req.TargetId); err != nil {
				return err
			}
			students, err = studentRepo.ListForUpdate(l.ctx, req.TargetId)
			if err != nil {
				return err
			}
			if len(students) == 0 {
				return &httperr.Error{Code: http.StatusBadRequest, Msg: "group has no students"}
			}
		case "class":
			students, err = studentRepo.ListForUpdate(l.ctx, 0)
			if err != nil {
				return err
			}
			if len(students) == 0 {
				return &httperr.Error{Code: http.StatusBadRequest, Msg: "class has no students"}
			}
		}

		now := time.Now()
		groupNames := make(map[int64]string)
		for i := range students {
			student := &students[i]
			groupName := ""
			if student.GroupID > 0 {
				if cached, ok := groupNames[student.GroupID]; ok {
					groupName = cached
				} else if group, getErr := groupRepo.Get(l.ctx, student.GroupID); getErr != nil {
					return getErr
				} else {
					groupName = group.Name
					groupNames[student.GroupID] = groupName
				}
			}
			entry := &model.ScoreEntry{
				StudentID:             student.ID,
				GroupID:               student.GroupID,
				DimensionID:           item.DimensionID,
				ScoreItemID:           item.ID,
				Score:                 item.Score,
				Remark:                req.Remark,
				RequestID:             requestID,
				StudentNoSnapshot:     student.StudentNo,
				StudentNameSnapshot:   student.Name,
				GroupNameSnapshot:     groupName,
				DimensionNameSnapshot: dimension.Name,
				ScoreItemNameSnapshot: item.Name,
			}
			if err := scoreEntryRepo.Create(l.ctx, entry); err != nil {
				return err
			}
			if err := studentRepo.AddScore(l.ctx, student.ID, item.Score); err != nil {
				return err
			}
			out = scoreEntryResp(entry)
		}
		if err := recentRepo.Touch(l.ctx, item.ID, now); err != nil {
			return err
		}
		return scoreOperationRepo.Complete(l.ctx, operation.ID, out.Id)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

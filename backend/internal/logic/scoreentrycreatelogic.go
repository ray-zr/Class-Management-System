// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"net/http"
	"time"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	it, err := l.svcCtx.ScoreItemRepo.Get(l.ctx, req.ScoreItemId)
	if err != nil {
		return nil, err
	}

	createForStudent := func(studentID int64) (*types.ScoreEntryResp, error) {
		s, err := l.svcCtx.StudentRepo.Get(l.ctx, studentID)
		if err != nil {
			return nil, err
		}
		e := &model.ScoreEntry{
			StudentID:   s.ID,
			GroupID:     s.GroupID,
			DimensionID: it.DimensionID,
			ScoreItemID: it.ID,
			Score:       it.Score,
			Remark:      req.Remark,
		}
		if err := l.svcCtx.ScoreEntryRepo.Create(l.ctx, e); err != nil {
			return nil, err
		}
		if err := l.svcCtx.RecentScoreItemRepo.Touch(l.ctx, it.ID, time.Now()); err != nil {
			return nil, err
		}
		newTotal := s.TotalScore + it.Score
		_, err = l.svcCtx.StudentRepo.Update(l.ctx, s.ID, map[string]any{"total_score": newTotal})
		if err != nil {
			return nil, err
		}
		return &types.ScoreEntryResp{
			Id:          e.ID,
			StudentId:   e.StudentID,
			GroupId:     e.GroupID,
			DimensionId: e.DimensionID,
			ScoreItemId: e.ScoreItemID,
			Score:       e.Score,
			Remark:      e.Remark,
			CreatedAt:   e.CreatedAt.Unix(),
		}, nil
	}

	switch req.Scope {
	case "student":
		if req.TargetId <= 0 {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing targetId"}
		}
		return createForStudent(req.TargetId)
	case "group":
		if req.TargetId <= 0 {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing targetId"}
		}
		_, err := l.svcCtx.GroupRepo.Get(l.ctx, req.TargetId)
		if err != nil {
			return nil, err
		}
		_, students, err := l.svcCtx.StudentRepo.List(l.ctx, repository.StudentListFilter{GroupID: req.TargetId})
		if err != nil {
			return nil, err
		}
		var last *types.ScoreEntryResp
		for _, s := range students {
			res, err := createForStudent(s.ID)
			if err != nil {
				return nil, err
			}
			last = res
		}
		if last == nil {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "group has no students"}
		}
		return last, nil
	case "class":
		_, students, err := l.svcCtx.StudentRepo.List(l.ctx, repository.StudentListFilter{})
		if err != nil {
			return nil, err
		}
		var last *types.ScoreEntryResp
		for _, s := range students {
			res, err := createForStudent(s.ID)
			if err != nil {
				return nil, err
			}
			last = res
		}
		if last == nil {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "class has no students"}
		}
		return last, nil
	default:
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid scope"}
	}
}

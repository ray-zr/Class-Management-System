// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RollcallPickLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRollcallPickLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RollcallPickLogic {
	return &RollcallPickLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RollcallPickLogic) RollcallPick() (resp *types.RollcallPickResp, err error) {
	roundID, fair, ok := l.svcCtx.RollcallState.Get()
	if !ok {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "rollcall not started"}
	}
	active, err := l.svcCtx.RollcallRepo.RoundActive(l.ctx, roundID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "rollcall ended"}
	}
	studentID, remaining, err := l.svcCtx.RollcallRepo.Pick(l.ctx, roundID, fair)
	if err != nil {
		if fair {
			_ = l.svcCtx.RollcallRepo.EndRound(l.ctx, roundID)
		}
		return nil, err
	}
	if fair && remaining == 0 {
		_ = l.svcCtx.RollcallRepo.EndRound(l.ctx, roundID)
	}
	s, err := l.svcCtx.StudentRepo.Get(l.ctx, studentID)
	if err != nil {
		return nil, err
	}
	return &types.RollcallPickResp{
		RoundId: roundID,
		Student: types.StudentResp{
			Id:         s.ID,
			StudentNo:  s.StudentNo,
			Name:       s.Name,
			Gender:     s.Gender,
			Phone:      s.Phone,
			Position:   s.Position,
			GroupId:    s.GroupID,
			TotalScore: s.TotalScore,
			CreatedAt:  s.CreatedAt.Unix(),
			UpdatedAt:  s.UpdatedAt.Unix(),
		},
		Remaining: remaining,
	}, nil
}

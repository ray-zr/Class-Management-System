// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type RollcallStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRollcallStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RollcallStartLogic {
	return &RollcallStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RollcallStartLogic) RollcallStart(req *types.RollcallStartReq) (resp *types.RollcallPickResp, err error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	roundID := uuid.NewString()
	fair := req.Fair
	if err := l.svcCtx.RollcallRepo.StartRound(l.ctx, roundID, fair); err != nil {
		return nil, err
	}
	l.svcCtx.RollcallState.Start(roundID, fair)
	studentID, remaining, err := l.svcCtx.RollcallRepo.Pick(l.ctx, roundID, fair)
	if err != nil {
		return nil, err
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

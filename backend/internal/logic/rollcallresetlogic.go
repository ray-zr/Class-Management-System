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

type RollcallResetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRollcallResetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RollcallResetLogic {
	return &RollcallResetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RollcallResetLogic) RollcallReset(req *types.RollcallResetReq) (resp *types.Empty, err error) {
	var roundID string
	if req != nil {
		roundID = req.RoundId
	}
	if roundID == "" {
		id, _, ok := l.svcCtx.RollcallState.Get()
		if !ok {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "rollcall not started"}
		}
		roundID = id
	}
	round, err := l.svcCtx.RollcallRepo.GetRound(l.ctx, roundID)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.RollcallRepo.Reset(l.ctx, roundID); err != nil {
		return nil, err
	}
	l.svcCtx.RollcallState.Start(roundID, round.Fair)
	return &types.Empty{}, nil
}

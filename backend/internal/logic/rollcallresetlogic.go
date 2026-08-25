// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"errors"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
		activeRound, activeErr := l.svcCtx.RollcallRepo.ActiveRound(l.ctx)
		if activeErr == nil {
			roundID = activeRound.RoundID
		} else if errors.Is(activeErr, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "rollcall not started"}
		} else {
			return nil, activeErr
		}
	}
	if err := l.svcCtx.RollcallRepo.Reset(l.ctx, roundID); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

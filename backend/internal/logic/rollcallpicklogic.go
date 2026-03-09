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

func (l *RollcallPickLogic) RollcallPick(req *types.RollcallPickReq) (resp *types.RollcallPickResp, err error) {
	roundID, fair, ok := l.svcCtx.RollcallState.Get()
	if !ok {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "rollcall not started"}
	}

	count := int64(1)
	if req != nil && req.Count > 0 {
		count = req.Count
	}
	if count < 1 {
		count = 1
	}
	if count > 50 {
		count = 50
	}

	active, err := l.svcCtx.RollcallRepo.RoundActive(l.ctx, roundID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "rollcall ended"}
	}

	items := make([]types.StudentResp, 0, int(count))
	remaining := int64(0)
	for i := int64(0); i < count; i++ {
		studentID, rem, pickErr := l.svcCtx.RollcallRepo.Pick(l.ctx, roundID, fair)
		if pickErr != nil {
			if fair {
				_ = l.svcCtx.RollcallRepo.EndRound(l.ctx, roundID)
			}
			if len(items) == 0 {
				return nil, pickErr
			}
			break
		}
		remaining = rem
		if fair && remaining == 0 {
			_ = l.svcCtx.RollcallRepo.EndRound(l.ctx, roundID)
		}
		s, getErr := l.svcCtx.StudentRepo.Get(l.ctx, studentID)
		if getErr != nil {
			return nil, getErr
		}
		items = append(items, types.StudentResp{
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
		})
		if fair && remaining == 0 {
			break
		}
	}

	var first *types.StudentResp
	if len(items) > 0 {
		first = &items[0]
	}
	return &types.RollcallPickResp{RoundId: roundID, Student: first, Students: items, Remaining: remaining}, nil
}

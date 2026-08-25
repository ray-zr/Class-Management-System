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
	round, err := l.svcCtx.RollcallRepo.ActiveRound(l.ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "rollcall not started"}
		}
		return nil, err
	}
	roundID, fair := round.RoundID, round.Fair

	requestedCount := int64(0)
	if req != nil {
		requestedCount = req.Count
	}
	count := normalizeRollcallCount(requestedCount)

	items := make([]types.StudentResp, 0, int(count))
	remaining := int64(0)
	for i := int64(0); i < count; i++ {
		studentID, rem, pickErr := l.svcCtx.RollcallRepo.Pick(l.ctx, roundID)
		if pickErr != nil {
			if len(items) == 0 {
				if errors.Is(pickErr, gorm.ErrRecordNotFound) {
					return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "rollcall ended"}
				}
				return nil, pickErr
			}
			break
		}
		remaining = rem
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

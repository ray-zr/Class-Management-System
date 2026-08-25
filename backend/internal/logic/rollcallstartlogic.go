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

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
	count := normalizeRollcallCount(req.Count)
	roundID := uuid.NewString()
	fair := req.Fair
	if err := l.svcCtx.RollcallRepo.StartRound(l.ctx, roundID, fair); err != nil {
		return nil, err
	}

	items := make([]types.StudentResp, 0, int(count))
	remaining := int64(0)
	for i := int64(0); i < count; i++ {
		studentID, rem, pickErr := l.svcCtx.RollcallRepo.Pick(l.ctx, roundID)
		if pickErr != nil {
			if len(items) == 0 {
				_ = l.svcCtx.RollcallRepo.EndRound(l.ctx, roundID)
				if errors.Is(pickErr, gorm.ErrRecordNotFound) {
					return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "class has no students"}
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

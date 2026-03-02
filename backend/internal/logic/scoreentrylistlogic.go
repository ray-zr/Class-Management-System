// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"net/http"
	"time"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScoreEntryListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreEntryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreEntryListLogic {
	return &ScoreEntryListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScoreEntryListLogic) ScoreEntryList(req *types.ScoreEntryListReq) (resp *types.ScoreEntryListResp, err error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	offset := (page - 1) * size
	limit := size
	sinceDays := req.SinceDays
	if sinceDays <= 0 {
		sinceDays = 30
	}
	since := time.Now().Add(-time.Duration(sinceDays) * 24 * time.Hour)

	total, items, err := l.svcCtx.ScoreEntryRepo.List(l.ctx, repository.ScoreEntryListFilter{
		StudentID: req.StudentId,
		GroupID:   req.GroupId,
		Since:     since,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	res := make([]types.ScoreEntryResp, 0, len(items))
	for _, e := range items {
		res = append(res, types.ScoreEntryResp{
			Id:          e.ID,
			StudentId:   e.StudentID,
			GroupId:     e.GroupID,
			DimensionId: e.DimensionID,
			ScoreItemId: e.ScoreItemID,
			Score:       e.Score,
			Remark:      e.Remark,
			CreatedAt:   e.CreatedAt.Unix(),
		})
	}
	return &types.ScoreEntryListResp{Total: total, Items: res}, nil
}

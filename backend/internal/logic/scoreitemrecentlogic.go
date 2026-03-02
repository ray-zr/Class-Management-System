// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScoreItemRecentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreItemRecentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreItemRecentLogic {
	return &ScoreItemRecentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScoreItemRecentLogic) ScoreItemRecent() (resp *types.RecentScoreItemsResp, err error) {
	ids, err := l.svcCtx.RecentScoreItemRepo.ListRecent(l.ctx, l.svcCtx.Config.App.RecentScoreItemsN)
	if err != nil {
		return nil, err
	}
	items := make([]types.ScoreItemResp, 0, len(ids))
	for _, id := range ids {
		it, err := l.svcCtx.ScoreItemRepo.Get(l.ctx, id)
		if err != nil {
			continue
		}
		items = append(items, types.ScoreItemResp{Id: it.ID, DimensionId: it.DimensionID, Name: it.Name, Score: it.Score})
	}
	return &types.RecentScoreItemsResp{Items: items}, nil
}

package logic

import (
	"context"
	"net/http"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

type GroupRankingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupRankingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupRankingLogic {
	return &GroupRankingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupRankingLogic) GroupRanking(req *types.RankingReq) (resp *types.GroupRankingResp, err error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	start, end, isTotal, rangeErr := rankingRange(time.Now(), req.Month, req.Total, req.Week)
	if rangeErr != nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: rangeErr.Error()}
	}

	var rows []repository.GroupScoreRow
	if isTotal {
		rows, err = l.svcCtx.RankingRepo.GroupTotalAvgScoreRanking(l.ctx)
	} else {
		rows, err = l.svcCtx.RankingRepo.GroupAvgFromStudentTotals(l.ctx, start, end, req.DimensionId)
	}
	if err != nil {
		return nil, err
	}

	topN := req.TopN
	if topN <= 0 {
		topN = l.svcCtx.Config.App.RankingTopN
	}
	if topN <= 0 {
		topN = 5
	}

	items := make([]types.GroupRankResp, 0, len(rows))
	var lastScoreVal float64
	var hasLast bool
	var rank int64
	var threshold float64
	var hasThreshold bool

	for _, row := range rows {
		score := row.Score
		if !hasLast || score != lastScoreVal {
			rank++
			lastScoreVal = score
			hasLast = true
			if !hasThreshold && rank >= topN {
				threshold = score
				hasThreshold = true
			}
		}
		highlight := false
		if topN > 0 {
			if !hasThreshold {
				highlight = true
			} else {
				highlight = score >= threshold
			}
		}
		items = append(items, types.GroupRankResp{
			Rank:      rank,
			Highlight: highlight,
			Group: types.GroupResp{
				Id:       row.GroupID,
				Name:     row.GroupName,
				AvgScore: row.Score,
			},
			Score: row.Score,
		})
	}

	return &types.GroupRankingResp{Items: items}, nil
}

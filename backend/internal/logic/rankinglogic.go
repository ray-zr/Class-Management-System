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

type RankingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRankingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RankingLogic {
	return &RankingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RankingLogic) Ranking(req *types.RankingReq) (resp *types.RankingResp, err error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	if req.Total {
		rows, err := l.svcCtx.RankingRepo.StudentTotalScoreRanking(l.ctx)
		if err != nil {
			return nil, err
		}
		topN := req.TopN
		if topN <= 0 {
			topN = l.svcCtx.Config.App.RankingTopN
		}
		return rankRows(rows, topN), nil
	}
	month := req.Month
	now := time.Now()
	if month == "" {
		month = now.Format("2006-01")
	}
	monthStart, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid month"}
	}
	monthEnd := monthStart.AddDate(0, 1, 0)
	rows, err := l.svcCtx.RankingRepo.StudentTotals(l.ctx, monthStart, monthEnd, req.DimensionId)
	if err != nil {
		return nil, err
	}
	topN := req.TopN
	if topN <= 0 {
		topN = l.svcCtx.Config.App.RankingTopN
	}
	return rankRows(rows, topN), nil
}

func rankRows(rows []repository.StudentScoreRow, topN int64) *types.RankingResp {
	if topN <= 0 {
		topN = 5
	}
	items := make([]types.StudentRankResp, 0, len(rows))
	var lastScoreVal int64
	var hasLastScore bool
	var rank int64
	var thresholdScore int64
	var hasThreshold bool
	for _, row := range rows {
		score := row.Score
		if !hasLastScore || score != lastScoreVal {
			rank++
			lastScoreVal = score
			hasLastScore = true
			if !hasThreshold && rank >= topN {
				thresholdScore = score
				hasThreshold = true
			}
		}
		highlight := false
		if topN > 0 {
			if !hasThreshold {
				highlight = true
			} else {
				highlight = score >= thresholdScore
			}
		}
		items = append(items, types.StudentRankResp{
			Rank:      rank,
			Highlight: highlight,
			Student: types.StudentResp{
				Id:         row.StudentID,
				StudentNo:  row.StudentNo,
				Name:       row.Name,
				Gender:     row.Gender,
				Phone:      row.Phone,
				Position:   row.Position,
				GroupId:    row.GroupID,
				TotalScore: row.TotalScore,
				CreatedAt:  row.StudentCreatedAt.Unix(),
				UpdatedAt:  row.StudentUpdatedAt.Unix(),
			},
			Score: score,
		})
	}
	return &types.RankingResp{Items: items}
}

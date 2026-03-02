// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

type RankingExportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRankingExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RankingExportLogic {
	return &RankingExportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RankingExportLogic) RankingExport(req *types.RankingReq) (file []byte, err error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	var rows []repository.StudentScoreRow
	if req.Total {
		rows, err = l.svcCtx.RankingRepo.StudentTotalScoreRanking(l.ctx)
	} else {
		month := req.Month
		if month == "" {
			month = time.Now().Format("2006-01")
		}
		monthStart, err := time.ParseInLocation("2006-01", month, time.Local)
		if err != nil {
			return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid month"}
		}
		monthEnd := monthStart.AddDate(0, 1, 0)
		rows, err = l.svcCtx.RankingRepo.StudentTotals(l.ctx, monthStart, monthEnd, req.DimensionId)
	}
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	sheetName := "月度积分排名汇总表"
	if req.Total {
		sheetName = "总分积分排名汇总表"
	}
	_ = f.SetSheetName(sheet, sheetName)
	sheet = sheetName
	_ = f.SetCellValue(sheet, "A1", "名次")
	_ = f.SetCellValue(sheet, "B1", "学号")
	_ = f.SetCellValue(sheet, "C1", "姓名")
	_ = f.SetCellValue(sheet, "D1", "小组")
	_ = f.SetCellValue(sheet, "E1", "职位")
	_ = f.SetCellValue(sheet, "F1", "积分")
	_ = f.SetCellValue(sheet, "G1", "奖惩标记")

	topN := req.TopN
	if topN <= 0 {
		topN = l.svcCtx.Config.App.RankingTopN
	}
	if topN <= 0 {
		topN = 5
	}

	var lastScoreVal int64
	var hasLastScore bool
	var rank int64
	var thresholdScore int64
	var hasThreshold bool
	rowNo := 2
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
		_ = f.SetCellValue(sheet, "A"+itoa(rowNo), rank)
		_ = f.SetCellValue(sheet, "B"+itoa(rowNo), row.StudentNo)
		_ = f.SetCellValue(sheet, "C"+itoa(rowNo), row.Name)
		_ = f.SetCellValue(sheet, "D"+itoa(rowNo), row.GroupName)
		_ = f.SetCellValue(sheet, "E"+itoa(rowNo), row.Position)
		_ = f.SetCellValue(sheet, "F"+itoa(rowNo), score)
		if highlight {
			_ = f.SetCellValue(sheet, "G"+itoa(rowNo), "奖")
		} else {
			_ = f.SetCellValue(sheet, "G"+itoa(rowNo), "")
		}
		if highlight {
			style, _ := f.NewStyle(&excelize.Style{Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFF2CC"}, Pattern: 1}})
			_ = f.SetCellStyle(sheet, "A"+itoa(rowNo), "G"+itoa(rowNo), style)
		}
		rowNo++
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return bytes.Clone(buf.Bytes()), nil
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

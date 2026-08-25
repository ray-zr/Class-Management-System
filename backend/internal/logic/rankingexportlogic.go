// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
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
	return &RankingExportLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RankingExportLogic) RankingExport(req *types.RankingReq) ([]byte, error) {
	if req == nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid request"}
	}
	start, end, isTotal, rangeErr := rankingRange(time.Now(), req.Month, req.StartDate, req.EndDate, req.Total)
	if rangeErr != nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: rangeErr.Error()}
	}

	rows, err := l.svcCtx.RankingRepo.StudentTotalScoreRanking(l.ctx)
	if err != nil {
		return nil, err
	}
	dimensions, err := l.svcCtx.DimensionRepo.List(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.DimensionId != 0 && !isTotal {
		filtered := dimensions[:0]
		for _, dimension := range dimensions {
			if dimension.ID == req.DimensionId {
				filtered = append(filtered, dimension)
			}
		}
		dimensions = filtered
	}
	entries, err := l.svcCtx.ScoreEntryRepo.ListCurrentStudentsForExport(l.ctx, start, end, exportDimensionID(req.DimensionId, isTotal))
	if err != nil {
		return nil, err
	}
	periodLabel := "全部历史记录"
	if !isTotal {
		periodLabel = fmt.Sprintf("%s 至 %s", start.Format("2006-01-02"), end.Add(-time.Nanosecond).Format("2006-01-02"))
	}
	return buildScoreExport(rows, dimensions, entries, periodLabel)
}

func exportDimensionID(dimensionID int64, total bool) int64 {
	if total {
		return 0
	}
	return dimensionID
}

func buildScoreExport(rows []repository.StudentScoreRow, dimensions []model.Dimension, entries []model.ScoreEntry, periodLabel string) ([]byte, error) {
	sort.Slice(rows, func(i, j int) bool { return studentRowLess(rows[i], rows[j]) })
	studentByID := make(map[int64]repository.StudentScoreRow, len(rows))
	for _, row := range rows {
		studentByID[row.StudentID] = row
	}
	sort.Slice(entries, func(i, j int) bool {
		left, right := studentByID[entries[i].StudentID], studentByID[entries[j].StudentID]
		if left.StudentNo != right.StudentNo || left.Name != right.Name {
			return studentRowLess(left, right)
		}
		if !entries[i].CreatedAt.Equal(entries[j].CreatedAt) {
			return entries[i].CreatedAt.Before(entries[j].CreatedAt)
		}
		return entries[i].ID < entries[j].ID
	})

	dimensionScores := make(map[int64]map[int64]int64, len(rows))
	addedScores := make(map[int64]int64, len(rows))
	deductedScores := make(map[int64]int64, len(rows))
	for _, entry := range entries {
		if dimensionScores[entry.StudentID] == nil {
			dimensionScores[entry.StudentID] = make(map[int64]int64, len(dimensions))
		}
		dimensionScores[entry.StudentID][entry.DimensionID] += entry.Score
		if entry.Score > 0 {
			addedScores[entry.StudentID] += entry.Score
		} else if entry.Score < 0 {
			deductedScores[entry.StudentID] += entry.Score
		}
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	const summarySheet = "量化汇总"
	const detailSheet = "积分细则"
	if err := f.SetSheetName(f.GetSheetName(0), summarySheet); err != nil {
		return nil, err
	}
	if _, err := f.NewSheet(detailSheet); err != nil {
		return nil, err
	}
	if err := writeSummarySheet(f, summarySheet, rows, dimensions, dimensionScores, addedScores, deductedScores, periodLabel); err != nil {
		return nil, err
	}
	if err := writeDetailSheet(f, detailSheet, entries, studentByID, periodLabel); err != nil {
		return nil, err
	}
	f.SetActiveSheet(0)
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return bytes.Clone(buf.Bytes()), nil
}

func writeSummarySheet(f *excelize.File, sheet string, rows []repository.StudentScoreRow, dimensions []model.Dimension, dimensionScores map[int64]map[int64]int64, addedScores, deductedScores map[int64]int64, periodLabel string) error {
	headers := []string{"序号", "学号", "姓名", "小组", "职位"}
	for _, dimension := range dimensions {
		headers = append(headers, dimension.Name)
	}
	headers = append(headers, "加分合计", "扣分合计", "期间净分", "当前总分")
	lastColumn, _ := excelize.ColumnNumberToName(len(headers))
	if err := writeExportHeading(f, sheet, lastColumn, "班级量化汇总表", "统计时间："+periodLabel+"；排序：学号、姓名", headers); err != nil {
		return err
	}
	for index, row := range rows {
		values := []any{index + 1, row.StudentNo, row.Name, row.GroupName, row.Position}
		for _, dimension := range dimensions {
			values = append(values, dimensionScores[row.StudentID][dimension.ID])
		}
		values = append(values, addedScores[row.StudentID], deductedScores[row.StudentID], addedScores[row.StudentID]+deductedScores[row.StudentID], row.TotalScore)
		if err := writeExportRow(f, sheet, index+5, values); err != nil {
			return err
		}
	}
	widths := []exportColumnWidth{{1, 1, 8}, {2, 2, 16}, {3, 5, 14}, {6, len(headers), 13}}
	return formatExportSheet(f, sheet, lastColumn, len(rows)+4, widths)
}

func writeDetailSheet(f *excelize.File, sheet string, entries []model.ScoreEntry, students map[int64]repository.StudentScoreRow, periodLabel string) error {
	headers := []string{"序号", "学号", "姓名", "小组", "记录时间", "积分维度", "积分细则", "分值", "备注"}
	if err := writeExportHeading(f, sheet, "I", "班级积分细则", "统计时间："+periodLabel+"；排序：学号、姓名、记录时间", headers); err != nil {
		return err
	}
	for index, entry := range entries {
		student := students[entry.StudentID]
		values := []any{index + 1, student.StudentNo, student.Name, student.GroupName, entry.CreatedAt.Format("2006-01-02 15:04:05"), entry.DimensionNameSnapshot, entry.ScoreItemNameSnapshot, entry.Score, entry.Remark}
		if err := writeExportRow(f, sheet, index+5, values); err != nil {
			return err
		}
	}
	widths := []exportColumnWidth{{1, 1, 8}, {2, 4, 16}, {5, 6, 20}, {7, 7, 36}, {8, 8, 10}, {9, 9, 32}}
	return formatExportSheet(f, sheet, "I", len(entries)+4, widths)
}

func writeExportHeading(f *excelize.File, sheet, lastColumn, title, subtitle string, headers []string) error {
	if err := f.MergeCell(sheet, "A1", lastColumn+"1"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A1", title); err != nil {
		return err
	}
	if err := f.MergeCell(sheet, "A2", lastColumn+"2"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", subtitle); err != nil {
		return err
	}
	for column, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(column+1, 4)
		if err := f.SetCellValue(sheet, cell, header); err != nil {
			return err
		}
	}
	return nil
}

func writeExportRow(f *excelize.File, sheet string, row int, values []any) error {
	for column, value := range values {
		cell, _ := excelize.CoordinatesToCellName(column+1, row)
		if column == 1 {
			if err := f.SetCellStr(sheet, cell, fmt.Sprint(value)); err != nil {
				return err
			}
			continue
		}
		if err := f.SetCellValue(sheet, cell, value); err != nil {
			return err
		}
	}
	return nil
}

type exportColumnWidth struct {
	min   int
	max   int
	width float64
}

func formatExportSheet(f *excelize.File, sheet, lastColumn string, lastRow int, widths []exportColumnWidth) error {
	titleStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#355C4D"}, Pattern: 1}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	if err != nil {
		return err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "#FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"#497A68"}, Pattern: 1}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}, Border: exportBorders()})
	if err != nil {
		return err
	}
	bodyStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Color: "#222222"}, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}, Border: exportBorders()})
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", lastColumn+"1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A4", lastColumn+"4", headerStyle); err != nil {
		return err
	}
	if lastRow >= 5 {
		if err := f.SetCellStyle(sheet, "A5", fmt.Sprintf("%s%d", lastColumn, lastRow), bodyStyle); err != nil {
			return err
		}
	}
	for _, width := range widths {
		minColumn, _ := excelize.ColumnNumberToName(width.min)
		maxColumn, _ := excelize.ColumnNumberToName(width.max)
		if err := f.SetColWidth(sheet, minColumn, maxColumn, width.width); err != nil {
			return err
		}
	}
	if err := f.SetRowHeight(sheet, 1, 28); err != nil {
		return err
	}
	if err := f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, YSplit: 4, TopLeftCell: "A5", ActivePane: "bottomLeft"}); err != nil {
		return err
	}
	return f.AutoFilter(sheet, fmt.Sprintf("A4:%s%d", lastColumn, lastRow), []excelize.AutoFilterOptions{})
}

func exportBorders() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: "#CBD5D1", Style: 1},
		{Type: "right", Color: "#CBD5D1", Style: 1},
		{Type: "top", Color: "#CBD5D1", Style: 1},
		{Type: "bottom", Color: "#CBD5D1", Style: 1},
	}
}

func studentRowLess(left, right repository.StudentScoreRow) bool {
	if left.StudentNo != right.StudentNo {
		leftNumber, leftNumeric := normalizedStudentNumber(left.StudentNo)
		rightNumber, rightNumeric := normalizedStudentNumber(right.StudentNo)
		if leftNumeric && rightNumeric {
			if len(leftNumber) != len(rightNumber) {
				return len(leftNumber) < len(rightNumber)
			}
			if leftNumber != rightNumber {
				return leftNumber < rightNumber
			}
		}
		return left.StudentNo < right.StudentNo
	}
	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.StudentID < right.StudentID
}

func normalizedStudentNumber(value string) (string, bool) {
	if value == "" {
		return "", false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return "", false
		}
	}
	normalized := strings.TrimLeft(value, "0")
	if normalized == "" {
		normalized = "0"
	}
	return normalized, true
}

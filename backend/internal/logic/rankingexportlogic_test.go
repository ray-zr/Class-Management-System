package logic

import (
	"bytes"
	"testing"
	"time"

	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/repository"

	"github.com/xuri/excelize/v2"
)

func TestBuildScoreExportOrdersStudentsAndIncludesDetails(t *testing.T) {
	rows := []repository.StudentScoreRow{
		{StudentID: 10, StudentNo: "10", Name: "十号", TotalScore: 3},
		{StudentID: 2, StudentNo: "2", Name: "二号", TotalScore: -1},
		{StudentID: 1, StudentNo: "1", Name: "一号", TotalScore: 0},
	}
	dimensions := []model.Dimension{
		{BaseModel: model.BaseModel{ID: 1}, Name: "纪律"},
		{BaseModel: model.BaseModel{ID: 2}, Name: "作业"},
	}
	now := time.Date(2026, 8, 25, 9, 30, 0, 0, time.Local)
	entries := []model.ScoreEntry{
		{BaseModel: model.BaseModel{ID: 2, CreatedAt: now.Add(time.Minute)}, StudentID: 10, DimensionID: 2, Score: 3, DimensionNameSnapshot: "作业", ScoreItemNameSnapshot: "作业全交"},
		{BaseModel: model.BaseModel{ID: 1, CreatedAt: now}, StudentID: 2, DimensionID: 1, Score: -1, DimensionNameSnapshot: "纪律", ScoreItemNameSnapshot: "迟到"},
	}

	data, err := buildScoreExport(rows, dimensions, entries, "2026-08-01 至 2026-08-25")
	if err != nil {
		t.Fatalf("buildScoreExport() error = %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open export: %v", err)
	}
	defer func() { _ = f.Close() }()

	if got := f.GetSheetList(); len(got) != 2 || got[0] != "量化汇总" || got[1] != "积分细则" {
		t.Fatalf("sheet list = %v", got)
	}
	for cell, want := range map[string]string{"B5": "1", "B6": "2", "B7": "10", "J6": "-1", "J7": "3"} {
		got, err := f.GetCellValue("量化汇总", cell)
		if err != nil || got != want {
			t.Fatalf("量化汇总 %s = %q, %v; want %q", cell, got, err, want)
		}
	}
	for cell, want := range map[string]string{"B5": "2", "G5": "迟到", "B6": "10", "G6": "作业全交"} {
		got, err := f.GetCellValue("积分细则", cell)
		if err != nil || got != want {
			t.Fatalf("积分细则 %s = %q, %v; want %q", cell, got, err, want)
		}
	}
}

func TestStudentRowLessUsesNaturalNumericOrder(t *testing.T) {
	one := repository.StudentScoreRow{StudentID: 1, StudentNo: "2", Name: "乙"}
	two := repository.StudentScoreRow{StudentID: 2, StudentNo: "10", Name: "甲"}
	if !studentRowLess(one, two) {
		t.Fatal("studentRowLess() did not place student number 2 before 10")
	}
}

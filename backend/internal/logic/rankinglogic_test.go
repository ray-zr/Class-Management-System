package logic

import (
	"testing"

	"class-management-system/backend/internal/repository"
)

func TestRankRowsUsesDenseRanksAndIncludesTies(t *testing.T) {
	rows := []repository.StudentScoreRow{
		{StudentID: 1, Score: 10},
		{StudentID: 2, Score: 10},
		{StudentID: 3, Score: 9},
		{StudentID: 4, Score: 8},
	}
	resp := rankRows(rows, 2)
	wantRanks := []int64{1, 1, 2, 3}
	wantHighlights := []bool{true, true, true, false}
	for i := range resp.Items {
		if resp.Items[i].Rank != wantRanks[i] {
			t.Errorf("item %d rank = %d, want %d", i, resp.Items[i].Rank, wantRanks[i])
		}
		if resp.Items[i].Highlight != wantHighlights[i] {
			t.Errorf("item %d highlight = %v, want %v", i, resp.Items[i].Highlight, wantHighlights[i])
		}
	}
}

func TestRankRowsIncludesScoreBreakdown(t *testing.T) {
	resp := rankRows([]repository.StudentScoreRow{{
		StudentID:     1,
		StudentNo:     "01",
		Name:          "测试学生",
		Score:         2,
		AddedScore:    5,
		DeductedScore: -3,
		EntryCount:    4,
	}}, 5)
	if len(resp.Items) != 1 {
		t.Fatalf("rankRows() item count = %d, want 1", len(resp.Items))
	}
	item := resp.Items[0]
	if item.AddedScore != 5 || item.DeductedScore != -3 || item.EntryCount != 4 {
		t.Fatalf("rankRows() breakdown = (%d, %d, %d)", item.AddedScore, item.DeductedScore, item.EntryCount)
	}
}

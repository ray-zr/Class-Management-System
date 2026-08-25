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

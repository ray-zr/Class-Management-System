package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

type RankingRepo struct{ db *gorm.DB }

func NewRankingRepo(db *gorm.DB) *RankingRepo { return &RankingRepo{db: db} }

type StudentScoreRow struct {
	StudentID     int64 `gorm:"column:student_id"`
	Score         int64 `gorm:"column:score"`
	AddedScore    int64 `gorm:"column:added_score"`
	DeductedScore int64 `gorm:"column:deducted_score"`
	EntryCount    int64 `gorm:"column:entry_count"`

	StudentNo  string `gorm:"column:student_no"`
	Name       string `gorm:"column:name"`
	Gender     string `gorm:"column:gender"`
	Phone      string `gorm:"column:phone"`
	Position   string `gorm:"column:position"`
	GroupID    int64  `gorm:"column:group_id"`
	GroupName  string `gorm:"column:group_name"`
	TotalScore int64  `gorm:"column:total_score"`

	StudentCreatedAt time.Time `gorm:"column:student_created_at"`
	StudentUpdatedAt time.Time `gorm:"column:student_updated_at"`
}

type GroupScoreRow struct {
	GroupID   int64   `gorm:"column:group_id"`
	GroupName string  `gorm:"column:group_name"`
	Score     float64 `gorm:"column:score"`
}

type GroupStudentScoreRow struct {
	GroupID   int64 `gorm:"column:group_id"`
	StudentID int64 `gorm:"column:student_id"`
	Score     int64 `gorm:"column:score"`
}

func (r *RankingRepo) StudentTotals(ctx context.Context, monthStart time.Time, monthEnd time.Time, dimensionID int64) ([]StudentScoreRow, error) {
	joinEntries := "LEFT JOIN score_entries e ON e.student_id = s.id AND e.created_at >= ? AND e.created_at < ?"
	joinArgs := []any{monthStart, monthEnd}
	if dimensionID != 0 {
		joinEntries = joinEntries + " AND e.dimension_id = ?"
		joinArgs = append(joinArgs, dimensionID)
	}

	selectSQL := fmt.Sprintf(
		"%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s",
		"s.id as student_id",
		"coalesce(sum(e.score), 0) as score",
		"coalesce(sum(case when e.score > 0 then e.score else 0 end), 0) as added_score",
		"coalesce(sum(case when e.score < 0 then e.score else 0 end), 0) as deducted_score",
		"count(e.id) as entry_count",
		"s.student_no as student_no",
		"s.name as name",
		"s.gender as gender",
		"s.phone as phone",
		"s.position as position",
		"s.group_id as group_id",
		"coalesce(g.name, '') as group_name",
		"s.total_score as total_score",
		"s.created_at as student_created_at",
		"s.updated_at as student_updated_at",
	)

	q := r.db.WithContext(ctx).
		Table("students s").
		Joins(joinEntries, joinArgs...).
		Joins("LEFT JOIN `groups` g ON g.id = s.group_id").
		Select(selectSQL).
		Group("s.id").
		Order("score desc, s.id asc")
	var res []StudentScoreRow
	if err := q.Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *RankingRepo) StudentTotalScoreRanking(ctx context.Context) ([]StudentScoreRow, error) {
	selectSQL := fmt.Sprintf(
		"%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s",
		"s.id as student_id",
		"s.total_score as score",
		"coalesce(es.added_score, 0) as added_score",
		"coalesce(es.deducted_score, 0) as deducted_score",
		"coalesce(es.entry_count, 0) as entry_count",
		"s.student_no as student_no",
		"s.name as name",
		"s.gender as gender",
		"s.phone as phone",
		"s.position as position",
		"s.group_id as group_id",
		"coalesce(g.name, '') as group_name",
		"s.total_score as total_score",
		"s.created_at as student_created_at",
		"s.updated_at as student_updated_at",
	)

	q := r.db.WithContext(ctx).
		Table("students s").
		Joins("LEFT JOIN `groups` g ON g.id = s.group_id").
		Joins(`LEFT JOIN (
			SELECT student_id,
			       SUM(CASE WHEN score > 0 THEN score ELSE 0 END) AS added_score,
			       SUM(CASE WHEN score < 0 THEN score ELSE 0 END) AS deducted_score,
			       COUNT(id) AS entry_count
			FROM score_entries
			GROUP BY student_id
		) es ON es.student_id = s.id`).
		Where("s.deleted_at IS NULL").
		Select(selectSQL).
		Order("score desc, s.id asc")
	var res []StudentScoreRow
	if err := q.Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *RankingRepo) GroupTotalAvgScoreRanking(ctx context.Context) ([]GroupScoreRow, error) {
	selectSQL := fmt.Sprintf(
		"%s, %s, %s",
		"g.id as group_id",
		"g.name as group_name",
		"coalesce(avg(s.total_score), 0) as score",
	)
	q := r.db.WithContext(ctx).
		Table("`groups` g").
		Joins("LEFT JOIN students s ON s.group_id = g.id AND s.deleted_at IS NULL").
		Select(selectSQL).
		Group("g.id").
		Order("score desc, g.id asc")
	var res []GroupScoreRow
	if err := q.Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *RankingRepo) GroupAvgFromStudentTotals(ctx context.Context, start time.Time, end time.Time, dimensionID int64) ([]GroupScoreRow, error) {
	rows, err := r.StudentTotals(ctx, start, end, dimensionID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []GroupScoreRow{}, nil
	}

	groupCnt := make(map[int64]int64, 16)
	groupSum := make(map[int64]int64, 16)
	groupName := make(map[int64]string, 16)
	for _, row := range rows {
		gid := row.GroupID
		if gid == 0 {
			continue
		}
		groupCnt[gid]++
		groupSum[gid] += row.Score
		if _, ok := groupName[gid]; !ok {
			groupName[gid] = row.GroupName
		}
	}

	res := make([]GroupScoreRow, 0, len(groupCnt))
	for gid, cnt := range groupCnt {
		sum := groupSum[gid]
		avg := float64(0)
		if cnt > 0 {
			avg = float64(sum) / float64(cnt)
		}
		res = append(res, GroupScoreRow{GroupID: gid, GroupName: groupName[gid], Score: avg})
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Score == res[j].Score {
			return res[i].GroupID < res[j].GroupID
		}
		return res[i].Score > res[j].Score
	})
	return res, nil
}

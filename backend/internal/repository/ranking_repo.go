package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type RankingRepo struct{ db *gorm.DB }

func NewRankingRepo(db *gorm.DB) *RankingRepo { return &RankingRepo{db: db} }

type StudentScoreRow struct {
	StudentID int64 `gorm:"column:student_id"`
	Score     int64 `gorm:"column:score"`

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

func (r *RankingRepo) StudentTotals(ctx context.Context, monthStart time.Time, monthEnd time.Time, dimensionID int64) ([]StudentScoreRow, error) {
	joinEntries := "LEFT JOIN score_entries e ON e.student_id = s.id AND e.created_at >= ? AND e.created_at < ?"
	joinArgs := []any{monthStart, monthEnd}
	if dimensionID != 0 {
		joinEntries = joinEntries + " AND e.dimension_id = ?"
		joinArgs = append(joinArgs, dimensionID)
	}

	selectSQL := fmt.Sprintf(
		"%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s",
		"s.id as student_id",
		"coalesce(sum(e.score), 0) as score",
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
		"%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s",
		"s.id as student_id",
		"s.total_score as score",
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
		Select(selectSQL).
		Order("score desc, s.id asc")
	var res []StudentScoreRow
	if err := q.Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

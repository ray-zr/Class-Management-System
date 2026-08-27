package db

import (
	"context"
	"fmt"
	"time"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
)

const (
	historyDimensionName = "历史数据校准"
	historyScoreItemName = "历史总分结转"
)

func PrepareCompatibilityMigrations(ctx context.Context, gdb *gorm.DB) error {
	if !gdb.WithContext(ctx).Migrator().HasTable(&model.RollcallPicked{}) {
		return nil
	}
	return gdb.WithContext(ctx).Exec(`
		DELETE duplicate
		FROM rollcall_picked duplicate
		JOIN rollcall_picked original
		  ON duplicate.round_id = original.round_id
		 AND duplicate.student_id = original.student_id
		 AND duplicate.id > original.id
	`).Error
}

func AutoMigrate(ctx context.Context, gdb *gorm.DB) error {
	return gdb.WithContext(ctx).AutoMigrate(
		&model.Group{},
		&model.Student{},
		&model.Dimension{},
		&model.ScoreItem{},
		&model.ScoreEntry{},
		&model.ScoreOperation{},
		&model.RecentScoreItem{},
		&model.RollcallRound{},
		&model.RollcallPicked{},
	)
}

func NormalizeGroupLayoutPositions(ctx context.Context, gdb *gorm.DB) error {
	return gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var groups []model.Group
		if err := tx.
			Order("CASE WHEN layout_position > 0 THEN 0 ELSE 1 END").
			Order("layout_position asc").
			Order("id asc").
			Find(&groups).Error; err != nil {
			return err
		}
		for index, group := range groups {
			position := int64(index + 1)
			if group.LayoutPosition == position {
				continue
			}
			if err := tx.Model(&model.Group{}).Where("id = ?", group.ID).Update("layout_position", position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func NormalizeStudentSeatPositions(ctx context.Context, gdb *gorm.DB) error {
	return gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var groupIDs []int64
		if err := tx.Model(&model.Student{}).Distinct("group_id").Order("group_id asc").Pluck("group_id", &groupIDs).Error; err != nil {
			return err
		}
		for _, groupID := range groupIDs {
			var students []model.Student
			if err := tx.Where("group_id = ?", groupID).
				Order("CASE WHEN seat_position > 0 THEN 0 ELSE 1 END").
				Order("seat_position asc").
				Order("id asc").
				Find(&students).Error; err != nil {
				return err
			}
			for index, student := range students {
				position := int64(index + 1)
				if student.SeatPosition == position {
					continue
				}
				if err := tx.Model(&model.Student{}).Where("id = ?", student.ID).Update("seat_position", position).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func BackfillScoreEntrySnapshots(ctx context.Context, gdb *gorm.DB) error {
	return gdb.WithContext(ctx).Exec(`
		UPDATE score_entries e
		LEFT JOIN students s ON s.id = e.student_id
		LEFT JOIN ` + "`groups`" + ` g ON g.id = e.group_id
		LEFT JOIN dimensions d ON d.id = e.dimension_id
		LEFT JOIN score_items si ON si.id = e.score_item_id
		SET
			e.student_no_snapshot = IF(e.student_no_snapshot = '', COALESCE(s.student_no, ''), e.student_no_snapshot),
			e.student_name_snapshot = IF(e.student_name_snapshot = '', COALESCE(s.name, ''), e.student_name_snapshot),
			e.group_name_snapshot = IF(e.group_name_snapshot = '', COALESCE(g.name, ''), e.group_name_snapshot),
			e.dimension_name_snapshot = IF(e.dimension_name_snapshot = '', COALESCE(d.name, ''), e.dimension_name_snapshot),
			e.score_item_name_snapshot = IF(e.score_item_name_snapshot = '', COALESCE(si.name, ''), e.score_item_name_snapshot)
		WHERE e.student_no_snapshot = ''
		   OR e.student_name_snapshot = ''
		   OR e.group_name_snapshot = ''
		   OR e.dimension_name_snapshot = ''
		   OR e.score_item_name_snapshot = ''
	`).Error
}

type scoreDifferenceRow struct {
	StudentID   int64  `gorm:"column:student_id"`
	GroupID     int64  `gorm:"column:group_id"`
	StudentNo   string `gorm:"column:student_no"`
	StudentName string `gorm:"column:student_name"`
	GroupName   string `gorm:"column:group_name"`
	Difference  int64  `gorm:"column:difference"`
}

// ReconcileStudentScoreDetails preserves legacy totals by recording any
// difference as an explicit historical carry-forward entry.
func ReconcileStudentScoreDetails(ctx context.Context, gdb *gorm.DB) error {
	return gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []scoreDifferenceRow
		if err := tx.Table("students s").
			Joins("LEFT JOIN `groups` g ON g.id = s.group_id").
			Joins("LEFT JOIN (SELECT student_id, SUM(score) AS entry_total FROM score_entries WHERE revoked_at IS NULL GROUP BY student_id) totals ON totals.student_id = s.id").
			Select(`s.id AS student_id,
				s.group_id AS group_id,
				s.student_no AS student_no,
				s.name AS student_name,
				COALESCE(g.name, '') AS group_name,
				s.total_score - COALESCE(totals.entry_total, 0) AS difference`).
			Where("s.total_score <> COALESCE(totals.entry_total, 0)").
			Scan(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}

		dimension := model.Dimension{Name: historyDimensionName}
		if err := tx.Where("name = ?", historyDimensionName).FirstOrCreate(&dimension).Error; err != nil {
			return err
		}
		item := model.ScoreItem{DimensionID: dimension.ID, Name: historyScoreItemName, Score: 1}
		if err := tx.Where("dimension_id = ? AND name = ?", dimension.ID, historyScoreItemName).FirstOrCreate(&item).Error; err != nil {
			return err
		}

		now := time.Now()
		for _, row := range rows {
			entry := model.ScoreEntry{
				StudentID:             row.StudentID,
				GroupID:               row.GroupID,
				DimensionID:           dimension.ID,
				ScoreItemID:           item.ID,
				Score:                 row.Difference,
				Remark:                "系统升级时按原总分与现有积分明细的差额自动结转",
				RequestID:             fmt.Sprintf("legacy-balance-%d-%d", row.StudentID, now.UnixNano()),
				StudentNoSnapshot:     row.StudentNo,
				StudentNameSnapshot:   row.StudentName,
				GroupNameSnapshot:     row.GroupName,
				DimensionNameSnapshot: historyDimensionName,
				ScoreItemNameSnapshot: historyScoreItemName,
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

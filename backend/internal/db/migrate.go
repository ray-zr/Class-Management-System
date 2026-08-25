package db

import (
	"context"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
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

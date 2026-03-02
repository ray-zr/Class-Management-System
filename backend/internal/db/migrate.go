package db

import (
	"context"

	"class-management-system/backend/internal/model"

	"gorm.io/gorm"
)

func AutoMigrate(ctx context.Context, gdb *gorm.DB) error {
	return gdb.WithContext(ctx).AutoMigrate(
		&model.Group{},
		&model.Student{},
		&model.Dimension{},
		&model.ScoreItem{},
		&model.ScoreEntry{},
		&model.RecentScoreItem{},
		&model.RollcallRound{},
		&model.RollcallPicked{},
	)
}

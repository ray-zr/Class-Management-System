// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"class-management-system/backend/internal/config"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/rollcall"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config              config.Config
	DB                  *gorm.DB
	StudentRepo         *repository.StudentRepo
	GroupRepo           *repository.GroupRepo
	DimensionRepo       *repository.DimensionRepo
	ScoreItemRepo       *repository.ScoreItemRepo
	RecentScoreItemRepo *repository.RecentScoreItemRepo
	ScoreEntryRepo      *repository.ScoreEntryRepo
	RankingRepo         *repository.RankingRepo
	RollcallRepo        *repository.RollcallRepo
	LockRepo            *repository.LockRepo
	RollcallState       *rollcall.State
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.Mysql.Dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:              c,
		DB:                  db,
		StudentRepo:         repository.NewStudentRepo(db),
		GroupRepo:           repository.NewGroupRepo(db),
		DimensionRepo:       repository.NewDimensionRepo(db),
		ScoreItemRepo:       repository.NewScoreItemRepo(db),
		RecentScoreItemRepo: repository.NewRecentScoreItemRepo(db),
		ScoreEntryRepo:      repository.NewScoreEntryRepo(db),
		RankingRepo:         repository.NewRankingRepo(db),
		RollcallRepo:        repository.NewRollcallRepo(db),
		LockRepo:            repository.NewLockRepo(db),
		RollcallState:       &rollcall.State{},
	}
}

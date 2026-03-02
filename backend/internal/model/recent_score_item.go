package model

type RecentScoreItem struct {
	BaseModel
	ScoreItemID int64 `gorm:"not null;uniqueIndex"`
	UsedAtUnix  int64 `gorm:"not null;index"`
}

func (RecentScoreItem) TableName() string { return "recent_score_items" }

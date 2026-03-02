package model

type ScoreEntry struct {
	BaseModel
	StudentID   int64  `gorm:"not null;index"`
	GroupID     int64  `gorm:"not null;index"`
	DimensionID int64  `gorm:"not null;index"`
	ScoreItemID int64  `gorm:"not null;index"`
	Score       int64  `gorm:"not null"`
	Remark      string `gorm:"type:varchar(255);not null;default:''"`
}

func (ScoreEntry) TableName() string { return "score_entries" }

package model

type ScoreItem struct {
	BaseModel
	DimensionID int64  `gorm:"not null;index"`
	Name        string `gorm:"type:varchar(128);not null"`
	Score       int64  `gorm:"not null"`
}

func (ScoreItem) TableName() string { return "score_items" }

package model

type ScoreEntry struct {
	BaseModel
	StudentID             int64  `gorm:"not null;index"`
	GroupID               int64  `gorm:"not null;index"`
	DimensionID           int64  `gorm:"not null;index"`
	ScoreItemID           int64  `gorm:"not null;index"`
	Score                 int64  `gorm:"not null"`
	Remark                string `gorm:"type:varchar(255);not null;default:''"`
	RequestID             string `gorm:"type:varchar(64);not null;default:'';index"`
	StudentNoSnapshot     string `gorm:"type:varchar(64);not null;default:''"`
	StudentNameSnapshot   string `gorm:"type:varchar(64);not null;default:''"`
	GroupNameSnapshot     string `gorm:"type:varchar(64);not null;default:''"`
	DimensionNameSnapshot string `gorm:"type:varchar(64);not null;default:''"`
	ScoreItemNameSnapshot string `gorm:"type:varchar(128);not null;default:''"`
}

func (ScoreEntry) TableName() string { return "score_entries" }

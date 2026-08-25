package model

type ScoreOperation struct {
	BaseModel
	RequestID   string `gorm:"type:varchar(64);not null;uniqueIndex"`
	Fingerprint string `gorm:"type:char(64);not null"`
	LastEntryID int64  `gorm:"not null;default:0"`
}

func (ScoreOperation) TableName() string { return "score_operations" }

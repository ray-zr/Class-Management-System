package model

type RollcallRound struct {
	BaseModel
	RoundID string `gorm:"type:varchar(64);not null;uniqueIndex"`
	Fair    bool   `gorm:"not null"`
	Active  bool   `gorm:"not null;index"`
}

func (RollcallRound) TableName() string { return "rollcall_rounds" }

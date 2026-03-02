package model

type RollcallPicked struct {
	BaseModel
	RoundID   string `gorm:"type:varchar(64);not null;index"`
	StudentID int64  `gorm:"not null;index"`
}

func (RollcallPicked) TableName() string { return "rollcall_picked" }

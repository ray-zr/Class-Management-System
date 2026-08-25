package model

type RollcallPicked struct {
	BaseModel
	RoundID   string `gorm:"type:varchar(64);not null;index;uniqueIndex:idx_rollcall_round_student"`
	StudentID int64  `gorm:"not null;index;uniqueIndex:idx_rollcall_round_student"`
}

func (RollcallPicked) TableName() string { return "rollcall_picked" }

package model

type Student struct {
	BaseModel
	StudentNo  string `gorm:"type:varchar(64);not null;uniqueIndex"`
	Name       string `gorm:"type:varchar(64);not null"`
	Gender     string `gorm:"type:varchar(16);not null"`
	Phone      string `gorm:"type:varchar(32);not null"`
	Position   string `gorm:"type:varchar(64);not null"`
	GroupID    int64  `gorm:"not null;index"`
	TotalScore int64  `gorm:"not null;default:0"`
}

func (Student) TableName() string { return "students" }

package model

import "gorm.io/gorm"

type Student struct {
	BaseModel
	StudentNo    string         `gorm:"type:varchar(64);not null;uniqueIndex"`
	Name         string         `gorm:"type:varchar(64);not null"`
	Gender       string         `gorm:"type:varchar(16);not null"`
	Phone        string         `gorm:"type:varchar(32);not null"`
	Position     string         `gorm:"type:varchar(64);not null"`
	GroupID      int64          `gorm:"not null;index;index:idx_students_group_seat,priority:1"`
	SeatPosition int64          `gorm:"not null;default:0;index:idx_students_group_seat,priority:2"`
	TotalScore   int64          `gorm:"not null;default:0"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (Student) TableName() string { return "students" }

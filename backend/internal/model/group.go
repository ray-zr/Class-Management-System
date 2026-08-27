package model

type Group struct {
	BaseModel
	Name           string `gorm:"type:varchar(64);not null;uniqueIndex"`
	LayoutPosition int64  `gorm:"not null;default:0;index"`
}

func (Group) TableName() string { return "groups" }

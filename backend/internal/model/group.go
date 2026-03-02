package model

type Group struct {
	BaseModel
	Name string `gorm:"type:varchar(64);not null;uniqueIndex"`
}

func (Group) TableName() string { return "groups" }

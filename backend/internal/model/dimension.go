package model

type Dimension struct {
	BaseModel
	Name string `gorm:"type:varchar(64);not null;uniqueIndex"`
}

func (Dimension) TableName() string { return "dimensions" }

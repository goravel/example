package models

import (
	"github.com/goravel/framework/database/orm"
)

type Category struct {
	orm.Timestamps
	ID   uint   `gorm:"primaryKey;column:id" json:"id"`
	Name string `gorm:"column:name" json:"name"`
}

func (r *Category) TableName() string {
	return "categories"
}

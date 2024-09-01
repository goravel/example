package models

import (
	"github.com/goravel/framework/database/orm"
)

type Restaurant struct {
	orm.Timestamps
	ID         uint       `gorm:"primaryKey;column:id" json:"id"`
	Name       string     `gorm:"column:name" json:"name"`
	Categories []Category `gorm:"many2many:restaurant_categories;joinForeignKey:restaurantId;joinReferences:categoryId" json:"categories"`
}

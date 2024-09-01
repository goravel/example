package models

import (
	"github.com/goravel/framework/database/orm"
)

type RestaurantCategory struct {
	orm.Model
	RestaurantID uint `gorm:"column:restaurantId"`
	CategoryID   uint `gorm:"column:categoryId"`
}

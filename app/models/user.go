package models

import (
	"github.com/goravel/framework/database/orm"
)

type User struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	Name   string
	Avatar string
	orm.SoftDeletes
	orm.Timestamps
}

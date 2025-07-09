package models

import (
	"github.com/goravel/framework/database/orm"
)

type User struct {
	orm.Model
	Name   string
	Avatar string
	Alias  string
	Mail   string
	orm.SoftDeletes
}

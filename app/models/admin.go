package models

import (
	"github.com/goravel/framework/database/orm"
)

type Admin struct {
	orm.Model
	Name       string  `json:"name" db:"name" form:"name"`
	Email      *string `json:"email" db:"email" form:"email"`
	Password   string  `json:"password" db:"password" form:"password"`
	CodeRendem *string `json:"code_rendem" db:"code_rendem" form:"code_rendem"`
	orm.SoftDeletes
}

func (r *Admin) TableName() string {
	return "admins"
}

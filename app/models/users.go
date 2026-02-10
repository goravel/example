package models

import "github.com/goravel/framework/database/orm"

type Users struct {
	orm.Model
	Name     string  `json:"name" db:"name"  form:"name"`
	Password string  `json:"password" db:"password"  form:"password"`
	Email    *string `json:"email" db:"email"  form:"email"`
	AdminId  uint64  `json:"admin_id" db:"admin_id" form:"admin_id"`
}

func (r *Users) TableName() string {
	return "users"
}


package models

import "github.com/goravel/framework/database/orm"

type Products struct {
	orm.Model
	NameProduct string  `json:"name_product" db:"name_product"`
	Description *string `json:"description" db:"description"`
	PathImg     *string `json:"path_img" db:"path_img"`
	AdminId     any     `json:"admin_id" db:"admin_id"`
}

func (r *Products) TableName() string {
	return "products"
}

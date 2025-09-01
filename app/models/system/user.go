package system

import (
	"github.com/goravel/framework/support/carbon"
)

type User struct {
	Id             int64            `json:"id" db:"id" gorm:"primaryKey"`
	Username       string           `json:"username" db:"username"`
	Password       *string          `json:"password" db:"password"`
	Nickname       string           `json:"nickname" db:"nickname"`
	OrganizationId int64            `json:"organizationId" db:"organizationId"`
	Avatar         *string          `json:"avatar" db:"avatar"`
	Sex            int8             `json:"sex" db:"sex"`
	Phone          *string          `json:"phone" db:"phone"`
	Email          *string          `json:"email" db:"email"`
	Card           *string          `json:"card" db:"card"`
	Birthday       *carbon.DateTime `json:"birthday" db:"birthday"`
	Introduction   *string          `json:"introduction" db:"introduction"`
	Estate         int8             `json:"estate" db:"estate"`
	CreateTime     carbon.DateTime  `json:"createTime" db:"createTime"`
	UpdateTime     carbon.DateTime  `json:"updateTime" db:"updateTime"`
	Inside         int8             `json:"inside" db:"inside"`
}

func (r *User) TableName() string {
	return "sys_user"
}

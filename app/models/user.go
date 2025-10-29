package models

import (
	"database/sql/driver"
	"errors"

	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/json"
)

type User struct {
	orm.Model
	Name   string
	Avatar string
	Alias  string
	Mail   string
	Tags   []UserTag `gorm:"serializer:json"`
	orm.SoftDeletes
}

type UserTag struct {
	Key string `json:"key"`
	Val int    `json:"value"`
}

func (r *UserTag) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, r)
}

func (r *UserTag) Value() (driver.Value, error) {
	return json.Marshal(r)
}

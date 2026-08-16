package models

import (
	"database/sql/driver"
	"errors"
	"strconv"

	"github.com/goravel/framework/contracts/notification"
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

// RouteNotificationFor resolves the delivery address per channel.
func (r *User) RouteNotificationFor(channel string) any {
	switch channel {
	case notification.ChannelMail:
		return r.Mail
	case notification.ChannelDatabase:
		return r.RouteNotificationForDatabase()
	default:
		return nil
	}
}

// RouteNotificationForDatabase implements contracts/notification.DatabaseRoutable:
// the type-safe database delivery route, preferred over RouteNotificationFor.
func (r *User) RouteNotificationForDatabase() string {
	return strconv.FormatUint(uint64(r.ID), 10)
}

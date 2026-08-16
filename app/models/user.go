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

// RouteNotificationForMail implements contracts/notification.MailRoutable:
// the type-safe mail delivery route, preferred over RouteNotificationFor.
func (r *User) RouteNotificationForMail(notification notification.Notification) map[string]string {
	return map[string]string{r.Mail: r.Name}
}

// RouteNotificationForDatabase implements contracts/notification.DatabaseRoutable:
// the type-safe database delivery route, preferred over RouteNotificationFor.
func (r *User) RouteNotificationForDatabase() string {
	return strconv.FormatUint(uint64(r.ID), 10)
}

// RouteNotificationFor resolves the delivery address per channel. The built-in
// mail and database channels resolve via the typed interfaces above, so no
// route is left to match on the channel name here.
func (r *User) RouteNotificationFor(channel string) any {
	return nil
}

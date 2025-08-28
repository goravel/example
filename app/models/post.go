package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

// Post is the blog post model.
type Post struct {
	orm.Model
	orm.SoftDeletes

	Title       string
	UserID      uint
	Slug        string
	Image       string
	Body        string
	PublishedAt time.Time
	User        *User
}

package models

import (
	"github.com/goravel/framework/database/orm"
)

type Post struct {
	orm.Model
	Title string `json:"title"`
	Body  string `json:"body"`
}

// TableName is prefixed with the module name so modules cannot collide.
func (r *Post) TableName() string {
	return "blog_posts"
}

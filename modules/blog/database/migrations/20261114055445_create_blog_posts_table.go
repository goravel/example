package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20261114055445CreateBlogPostsTable struct{}

// Signature The unique signature for the migration.
func (r *M20261114055445CreateBlogPostsTable) Signature() string {
	return "20261114055445_create_blog_posts_table"
}

// Up Run the migrations.
func (r *M20261114055445CreateBlogPostsTable) Up() error {
	if !facades.Schema().HasTable("blog_posts") {
		return facades.Schema().Create("blog_posts", func(table schema.Blueprint) {
			table.ID()
			table.String("title")
			table.Text("body")
			table.TimestampsTz()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20261114055445CreateBlogPostsTable) Down() error {
	return facades.Schema().DropIfExists("blog_posts")
}

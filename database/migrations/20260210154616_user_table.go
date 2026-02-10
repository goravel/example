package migrations

import (
	"goravel/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260210154616UserTable struct{}

// Signature The unique signature for the migration.
func (r *M20260210154616UserTable) Signature() string {
	return "20260210154616_user_table"
}

// Up Run the migrations.
func (r *M20260210154616UserTable) Up() error {

	if !facades.Schema().HasTable("users") {

		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.ID().AutoIncrement()
			table.String("name")
			table.String("password")
			table.String("email").Nullable()

			// forenkey ke table admin
			table.UnsignedBigInteger("admin_id")
			table.Foreign("admin_id").References("id").On("admins").CascadeOnDelete()

			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("update_at").UseCurrent()

		})
	}
	return nil

}

// Down Reverse the migrations.
func (r *M20260210154616UserTable) Down() error {
	return facades.Schema().DropIfExists("users")
}


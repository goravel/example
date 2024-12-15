package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20241207095921CreateUsersTable struct {
}

// Signature The unique signature for the migration.
func (r *M20241207095921CreateUsersTable) Signature() string {
	return "20241207095921_create_users_table"
}

// Up Run the migrations.
func (r *M20241207095921CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("name").Default("")
			table.String("avatar").Default("")
			table.Timestamps()
			table.SoftDeletes()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20241207095921CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}

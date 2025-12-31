package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20250330911908AddColumnsToUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M20250330911908AddColumnsToUsersTable) Signature() string {
	return "20250331111908_add_columns_to_users_table"
}

// Up Run the migrations.
func (r *M20250330911908AddColumnsToUsersTable) Up() error {
	return facades.Schema().Table("users", func(table schema.Blueprint) {
		table.String("alias").Default("").After("name")
		table.String("email").Nullable().First()
	})
}

// Down Reverse the migrations.
func (r *M20250330911908AddColumnsToUsersTable) Down() error {
	return nil
}

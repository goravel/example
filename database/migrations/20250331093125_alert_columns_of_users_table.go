package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250331093125AlertColumnsOfUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M20250331093125AlertColumnsOfUsersTable) Signature() string {
	return "20250331093125_alert_columns_of_users_table"
}

// Up Run the migrations.
func (r *M20250331093125AlertColumnsOfUsersTable) Up() error {
	if facades.Schema().HasTable("users") {
		return facades.Schema().Table("users", func(table schema.Blueprint) {
			table.String("alias").Default("test").Change()
			table.RenameColumn("email", "mail")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20250331093125AlertColumnsOfUsersTable) Down() error {
	return nil
}

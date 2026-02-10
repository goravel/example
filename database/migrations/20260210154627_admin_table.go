package migrations

import (
	"goravel/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260210154627AdminTable struct{}

// Signature The unique signature for the migration.
func (r *M20260210154627AdminTable) Signature() string {
	return "20260210154627_admin_table"
}

// Up Run the migrations.
func (r *M20260210154627AdminTable) Up() error {
	if !facades.Schema().HasTable("admins") {
		return facades.Schema().Create("admins", func(table schema.Blueprint) {
			table.ID().AutoIncrement()
			table.String("name")
			table.String("email").Nullable()
			table.String("password")
			table.String("code_rendem").Nullable() // misalnya nantik admin di batasi menerima user sebelum di tambah code nya di settings

			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("update_at").UseCurrent()
		})
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260210154627AdminTable) Down() error {

	return facades.Schema().DropIfExists("admins")

}

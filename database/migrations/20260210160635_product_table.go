package migrations

import (
	"goravel/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260210160635ProductTable struct{}

// Signature The unique signature for the migration.
func (r *M20260210160635ProductTable) Signature() string {
	return "20260210160635_product_table"
}

// Up Run the migrations.
func (r *M20260210160635ProductTable) Up() error {
	if !facades.Schema().HasTable("products") {
		return facades.Schema().Create("products", func(table schema.Blueprint) {
			table.ID().AutoIncrement()
			table.String("name_product")
			table.String("description").Nullable()
			table.String("path_img").Nullable()
			table.UnsignedBigInteger("admin_id")
			table.Foreign("admin_id").References("id").On("admins").CascadeOnDelete().CascadeOnDelete().CascadeOnUpdate()

		})
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260210160635ProductTable) Down() error {
	return nil
}

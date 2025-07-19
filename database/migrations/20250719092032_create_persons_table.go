package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250719092032CreatePersonsTable struct{}

// Signature The unique signature for the migration.
func (r *M20250719092032CreatePersonsTable) Signature() string {
	return "20250719092032_create_persons_table"
}

// Up Run the migrations.
func (r *M20250719092032CreatePersonsTable) Up() error {
	if !facades.Schema().HasTable("persons") {
		return facades.Schema().Create("persons", func(table schema.Blueprint) {
			table.ID()
			table.Integer("type").Nullable().Comment("人员类型")
			table.Integer("sex").Nullable().Comment("性别")
			table.Integer("education").Nullable().Comment("学历 专科 本科 本科以上")
			table.Integer("graduate").Nullable().Comment("在学校状态")
			table.String("street_code", 100).Nullable().Comment("街道编码")

			table.Index("type").Name("type")
			table.Index("sex").Name("sex")
			table.Index("education").Name("education")
			table.Index("graduate").Name("graduate")
			table.Index("street_code").Name("street_code")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20250719092032CreatePersonsTable) Down() error {
	return facades.Schema().DropIfExists("persons")
}

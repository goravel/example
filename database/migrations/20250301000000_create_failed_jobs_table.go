package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250301000000CreateFailedJobsTable struct{}

// Signature The unique signature for the migration.
func (r *M20250301000000CreateFailedJobsTable) Signature() string {
	return "20250301000000_create_failed_jobs_table"
}

// Up Run the migrations.
func (r *M20250301000000CreateFailedJobsTable) Up() error {
	if !facades.Schema().HasTable("failed_jobs") {
		return facades.Schema().Create("failed_jobs", func(table schema.Blueprint) {
			table.ID()
			table.String("uuid")
			table.Unique("uuid")
			table.Text("connection")
			table.Text("queue")
			table.LongText("payload")
			table.LongText("exception")
			table.DateTime("failed_at").UseCurrent()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20250301000000CreateFailedJobsTable) Down() error {
	return facades.Schema().DropIfExists("failed_jobs")
}

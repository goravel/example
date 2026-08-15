package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260805000001CreateNotificationsTable struct{}

func (r *M20260805000001CreateNotificationsTable) Signature() string {
	return "20260805000001_create_notifications_table"
}

func (r *M20260805000001CreateNotificationsTable) Up() error {
	if facades.Schema().HasTable("notifications") {
		return nil
	}

	return facades.Schema().Create("notifications", func(table schema.Blueprint) {
		table.String("id", 36)
		table.Primary("id")
		table.String("type")
		table.String("notifiable_type")
		table.String("notifiable_id")
		table.Text("data")
		table.Timestamp("read_at").Nullable()
		table.Timestamps()
		table.Index("notifiable_type", "notifiable_id")
	})
}

func (r *M20260805000001CreateNotificationsTable) Down() error {
	return facades.Schema().DropIfExists("notifications")
}

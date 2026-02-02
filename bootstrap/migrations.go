package bootstrap

import (
	"goravel/database/migrations"

	"github.com/goravel/framework/contracts/database/schema"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateUsersTable{},
		&migrations.M20210101000002CreateJobsTable{},
		&migrations.M20250330911908AddColumnsToUsersTable{},
		&migrations.M20250331093125AlertColumnsOfUsersTable{},
	}
}

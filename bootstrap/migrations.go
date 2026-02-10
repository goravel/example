package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20260210154616UserTable{},
		&migrations.M20260210154627AdminTable{},
		&migrations.M20260210160635ProductTable{},
	}
}

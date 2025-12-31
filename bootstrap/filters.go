package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/filters"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&filters.Test{},
	}
}

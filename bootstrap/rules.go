package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.Exists{},
		&rules.NotExists{},
	}
}

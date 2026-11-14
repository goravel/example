package bootstrap

import (
	"github.com/goravel/framework/contracts/modular"

	"goravel/modules/admin"
	"goravel/modules/blog"
)

func Modules() []modular.Module {
	return []modular.Module{
		&admin.Module{},
		&blog.Module{},
	}
}

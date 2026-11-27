package bootstrap

import (
	"github.com/goravel/framework/contracts/module"

	"goravel/modules/admin"
	"goravel/modules/blog"
)

func Modules() []module.Module {
	return []module.Module{
		&admin.Module{},
		&blog.Module{},
	}
}

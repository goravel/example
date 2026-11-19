package facades

import (
	"github.com/goravel/framework/contracts/module"
)

func Module() module.Registry {
	return App().MakeModule()
}

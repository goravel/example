package facades

import (
	"github.com/goravel/framework/contracts/modular"
)

func Modular() modular.Registry {
	return App().MakeModular()
}

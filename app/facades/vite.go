package facades

import (
	"github.com/goravel/framework/contracts/vite"
)

func Vite() vite.Vite {
	return App().MakeVite()
}

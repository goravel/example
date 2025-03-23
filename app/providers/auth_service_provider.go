package providers

import (
	framewrokauth "github.com/goravel/framework/auth"
	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
)

type AuthServiceProvider struct {
}

func (receiver *AuthServiceProvider) Register(app foundation.Application) {

}

func (receiver *AuthServiceProvider) Boot(app foundation.Application) {
	facades.Auth().Extend("session", func(name string, auth contractsauth.Auth, userProvider contractsauth.UserProvider) (contractsauth.GuardDriver, error) {
		return framewrokauth.NewJwtGuard(app.Request(), name, facades.Cache(), facades.Config(), userProvider)
	})
}

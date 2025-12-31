package providers

import (
	"goravel/app/facades"

	frameworkauth "github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
)

type AuthServiceProvider struct {
}

func (receiver *AuthServiceProvider) Register(app foundation.Application) {

}

func (receiver *AuthServiceProvider) Boot(app foundation.Application) {
	facades.Auth().Extend("another-jwt", frameworkauth.NewJwtGuard)
	facades.Auth().Provider("another-orm", frameworkauth.NewOrmUserProvider)
}

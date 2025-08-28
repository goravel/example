package providers

import (
	"context"

	"goravel/app/models"

	frameworkauth "github.com/goravel/framework/auth"
	"github.com/goravel/framework/auth/access"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
)

type AuthServiceProvider struct {
}

func (receiver *AuthServiceProvider) Register(app foundation.Application) {

}

func (receiver *AuthServiceProvider) Boot(app foundation.Application) {
	facades.Auth().Extend("another-jwt", frameworkauth.NewJwtGuard)
	facades.Auth().Provider("another-orm", frameworkauth.NewOrmUserProvider)
	facades.Gate().Define("update-post", func(ctx context.Context, arguments map[string]any) contractsaccess.Response {
		user := ctx.Value("user").(models.User)

		post := arguments["post"].(models.Post)

		if user.ID == post.UserID {
			return access.NewAllowResponse()
		} else {
			return access.NewDenyResponse("error")
		}
	})
}

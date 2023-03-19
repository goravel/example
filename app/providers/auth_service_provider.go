package providers

import (
	"context"

	"github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/app/policies"
)

type AuthServiceProvider struct {
}

func (receiver *AuthServiceProvider) Register() {

}

func (receiver *AuthServiceProvider) Boot() {
	facades.Gate.Before(func(ctx context.Context, ability string, arguments map[string]any) access.Response {
		//user := ctx.Value("user").(models.User)
		//if isAdministrator(user) {
		//	return access.NewAllowResponse()
		//}

		return nil
	})

	facades.Gate.After(func(ctx context.Context, ability string, arguments map[string]any, result access.Response) access.Response {
		//user := ctx.Value("user").(models.User)
		//if isAdministrator(user) {
		//	return access.NewAllowResponse()
		//}

		return nil
	})

	facades.Gate.Define("update-user", func(ctx context.Context, arguments map[string]any) access.Response {
		//user := ctx.Value("user").(models.User)
		user := arguments["user"].(models.User)
		userID := arguments["userID"].(uint)

		if user.ID == userID {
			return access.NewAllowResponse()
		} else {
			return access.NewDenyResponse("cannot update")
		}
	})

	facades.Gate.Define("create-user", policies.NewUserPolicy().Create)
}

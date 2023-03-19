package policies

import (
	"context"

	"github.com/goravel/framework/contracts/auth/access"

	"goravel/app/models"
)

type UserPolicy struct {
}

func NewUserPolicy() *UserPolicy {
	return &UserPolicy{}
}

func (r *UserPolicy) Create(ctx context.Context, arguments map[string]any) access.Response {
	//user := ctx.Value("user").(models.User)
	user := arguments["user"].(models.User)
	userID := arguments["userID"].(uint)

	if user.ID == userID {
		return access.NewAllowResponse()
	} else {
		return access.NewDenyResponse("cannot create")
	}
}

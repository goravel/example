package controllers

import (
	"fmt"

	proto "github.com/goravel/example-proto"
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

/*********************************
gRPC Client Example

This is the gRPC Client side example, if you need the full steps about gRPC, please visit the link below.
https://github.com/goravel/example/blob/master/app/grpc/controllers/user_controller.go
********************************/

type GrpcController struct {
	userService proto.UserServiceClient
}

func NewGrpcController() *GrpcController {
	// The initialization process can be moved to app/services/*.go
	client, err := facades.Grpc().Connect("user")
	if err != nil {
		facades.Log().Error(fmt.Sprintf("failed to connect to user server: %+v", err))
	}

	return &GrpcController{
		userService: proto.NewUserServiceClient(client),
	}
}

func (r *GrpcController) User(ctx http.Context) http.Response {
	resp, err := r.userService.GetUser(ctx, &proto.UserRequest{
		Token: ctx.Request().Input("token"),
	})
	if err != nil {
		return ctx.Response().String(http.StatusInternalServerError, fmt.Sprintf("call UserService.GetUser err: %+v", err))
	}
	if resp.Code != http.StatusOK {
		return ctx.Response().String(http.StatusInternalServerError, fmt.Sprintf("user service returns error, code: %d, message: %s", resp.Code, resp.Message))
	}

	return ctx.Response().Success().Json(resp.GetData())
}

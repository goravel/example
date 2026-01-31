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

type GrpcController struct{}

func NewGrpcController() *GrpcController {
	return &GrpcController{}
}

func (r *GrpcController) User(ctx http.Context) http.Response {
	client, err := facades.Grpc().Client(ctx, "user")
	if err != nil {
		return ctx.Response().String(http.StatusInternalServerError, fmt.Sprintf("init UserService err: %+v", err))
	}

	userServiceClient := proto.NewUserServiceClient(client)
	resp, err := userServiceClient.GetUser(ctx, &proto.UserRequest{
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

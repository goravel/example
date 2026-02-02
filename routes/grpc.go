package routes

import (
	proto "github.com/goravel/example-proto"

	"goravel/app/facades"
	"goravel/app/grpc/controllers"
	httpcontrollers "goravel/app/http/controllers"
)

func Grpc() {
	proto.RegisterUserServiceServer(facades.Grpc().Server(), controllers.NewUserController())

	grpcController := httpcontrollers.NewGrpcController()
	facades.Route().Get("/grpc/user", grpcController.User)
}

package routes

import (
	"goravel/app/facades"

	proto "github.com/goravel/example-proto"

	"goravel/app/grpc/controllers"
)

func Grpc() {
	proto.RegisterUserServiceServer(facades.Grpc().Server(), controllers.NewUserController())
}

package routes

import (
	proto "github.com/goravel/example-proto"

	"goravel/app/facades"
	"goravel/app/grpc/controllers"
)

func Grpc() {
	proto.RegisterUserServiceServer(facades.Grpc().Server(), controllers.NewUserController())
}

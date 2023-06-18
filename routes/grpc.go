package routes

import (
	proto "github.com/goravel/example-proto"
	"github.com/goravel/framework/facades"

	"goravel/app/grpc/controllers"
)

func Grpc() {
	proto.RegisterUserServiceServer(facades.Grpc().Server(), controllers.NewUserController())
}

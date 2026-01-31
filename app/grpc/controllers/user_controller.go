package controllers

import (
	"context"
	"net/http"

	proto "github.com/goravel/example-proto"
)

/*********************************

#gRPC Example

## Server Side

1. Configure gRPC host in the .env file

GRPC_HOST=127.0.0.1
GRPC_PORT=3001

2. Get gRPC proto, you can publish your own proto to Github and get it like below.

`go get github.com/goravel/example-proto`

3. Add get user logic in `app/grpc/controllers/user_controller.go`

4. Add route in `route/grpc.go`

`proto.RegisterUserServiceServer(facades.Grpc().Server(), controllers.NewUserController())`

## Client Side

The client side is defined in this project as well for simplicity, you can also create a new Goravel project as the client side.

1. Configure gRPC in .env file(goravel/example-client)

GRPC_USER_HOST=127.0.0.1
GRPC_USER_PORT=3001

2. Add gRPC connection to config/grpc.go

```
"connections": map[string]any{
	"user": map[string]any{
		"host":         config.Env("GRPC_USER_HOST", ""),
		"port":         config.Env("GRPC_USER_PORT", ""),
		"interceptors": []string{},
	},
},
```

3. Add `app/http/controllers/user_controller.go` to call the server side(goravel/example-client)

4. Add get user route to route/grpc.go

```
grpcController := httpcontrollers.NewGrpcController()
facades.Route().Get("/grpc/user", grpcController.User)
```

## Run Server and Client

`air`

## Test

Make a request to the client side, the client side will call the server side by gRPC to get user info.

```
curl --location --request GET 'http://127.0.0.1:3010/user?token=1'
```
********************************/

type UserController struct {
}

func NewUserController() *UserController {
	return &UserController{}
}

func (receiver *UserController) GetUser(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	return &proto.UserResponse{
		Code: http.StatusOK,
		Data: &proto.User{
			Id:    1,
			Name:  "Goravel",
			Token: req.GetToken(),
		},
	}, nil
}

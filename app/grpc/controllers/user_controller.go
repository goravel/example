package controllers

import (
	"context"
	"net/http"

	proto "github.com/goravel/example-proto"
	"github.com/goravel/framework/auth"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	goravelhttp "github.com/goravel/framework/http"
	"github.com/pkg/errors"

	"goravel/app/models"
)

/*********************************
gRPC Example

This is an example for gRPC. There are a server and a client.
This repository is the Server side, it provides get user by token function.
The repository of Client is `git@github.com:goravel/example-client.git`,
The Client can get user by token from the server.
You need initialize your JWT, DB first in this repository, you can find the step in
`app/http/controllers/jwt_controller.go` and `app/http/controllers/db_controller.go`
[gRPC Document](https://www.goravel.dev/the-basics/grpc.html)

1. Configure gRPC host in the .env file
GRPC_HOST=127.0.0.1
GRPC_PORT=3001

2. Get gRPC proto, you can publish your own proto to Github and get it like below.
`go get github.com/goravel/example-proto`

3. Add get user logic to `app/grpc/controllers/user_controller.go`

4. Add route to `route/grpc.go`

5. Add gRPC Server in `main.go`
```
go func() {
	if err := facades.Grpc().Run(); err != nil {
		facades.Log().Errorf("Run grpc error: %+v", err)
	}
}()
```

6. Run Server
`air`

6. Clone gRPC Client
```
cd .. && git clone git@github.com:goravel/example-client.git && cd example-client
```

7. Configure gRPC in .env file(goravel/example-client)
GRPC_USER_HOST=127.0.0.1
GRPC_USER_PORT=3001

8. Add gRPC client to config/grpc.go file(goravel/example-client)
```
"clients": map[string]any{
	"user": map[string]any{
		"host":         config.Env("GRPC_USER_HOST", ""),
		"port":         config.Env("GRPC_USER_PORT", ""),
		"interceptors": []string{},
	},
},
```

9. Add User Service to `app/services/user`, it is used to call the server side(goravel/example-client)

10. Add `app/http/controllers/user_controller.go` to call UserService(goravel/example-client)

11. Add get user route to `route/web.go`(goravel/example-client)

13. Run Client(goravel/example-client)
`air`

14. Get Token that used by Client
```
curl --location --request GET 'http://127.0.0.1:3000/jwt/login'
```

15. Get user from Server by calling `http://127.0.0.1:3010/user` that through Client
```
curl --location --request GET 'http://127.0.0.1:3010/user' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJrZXkiOiIxIiwic3ViIjoidXNlciIsImV4cCI6MTY3OTIxNzUwNSwiaWF0IjoxNjc5MjEzOTA1fQ.SH32-ZImkWw4zMdYPokSNRR8TMZp4vD5c2ZO4sTbk_0'
```

16. We succeed get user by gRPC from Server in Client

17. There is a server interceptor example about opentracing, you can find it in `app/grpc/interceptors/opentracing_server.go`
 ********************************/

type UserController struct {
}

func NewUserController() *UserController {
	return &UserController{}
}

func (receiver *UserController) GetUser(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	if req.GetToken() == "" {
		return &proto.UserResponse{
			Code:    http.StatusUnauthorized,
			Message: "empty token",
		}, nil
	}

	httpCtx := goravelhttp.Background()
	token, err := refreshToken(httpCtx, req.GetToken())
	if err != nil {
		return &proto.UserResponse{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		}, nil
	}

	user, err := getUser(httpCtx)
	if err != nil {
		return &proto.UserResponse{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		}, nil
	}

	return &proto.UserResponse{
		Code: http.StatusOK,
		Data: &proto.User{
			Id:        uint64(user.ID),
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
			Name:      user.Name,
			Avatar:    user.Avatar,
			Token:     token,
		},
	}, nil
}

func refreshToken(ctx contractshttp.Context, token string) (string, error) {
	if _, err := facades.Auth().Parse(ctx, token); err != nil {
		if errors.Is(err, auth.ErrorTokenExpired) {
			token, err = facades.Auth().Refresh(ctx)
			if err != nil {
				return "", err
			}

			token = "Bearer " + token
		} else {
			return "", err
		}
	}

	return token, nil
}

func getUser(ctx contractshttp.Context) (models.User, error) {
	var user models.User
	if err := facades.Auth().User(ctx, &user); err != nil {
		return user, err
	}

	if user.ID == 0 {
		return user, errors.New("no user found")
	}

	return user, nil
}

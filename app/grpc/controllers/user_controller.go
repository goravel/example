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

type UserController struct {
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
	if _, err := facades.Auth.Parse(ctx, token); err != nil {
		if errors.Is(err, auth.ErrorTokenExpired) {
			token, err = facades.Auth.Refresh(ctx)
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
	if err := facades.Auth.User(ctx, &user); err != nil {
		return user, err
	}

	if user.ID == 0 {
		return user, errors.New("no user found")
	}

	return user, nil
}

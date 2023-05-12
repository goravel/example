package controllers

import (
	"github.com/goravel/framework/contracts/http"
)

/*********************************
1. Install swag
document: https://github.com/swaggo/http-swagger

go install github.com/swaggo/swag/cmd/swag@latest

2. Init document
swag init

3. Install http-swagger
go get -u github.com/swaggo/http-swagger

4. Optimize the document of endpoint: app/http/controllers/swagger_controller.go

5. Add route to `/route/web.go`

6. Run Server
air

7. Visit: http://localhost:3000/swagger/index.html
 ********************************/

type SwaggerController struct {
	//Dependent services
}

func NewSwaggerController() *JwtController {
	return &JwtController{
		//Inject services
	}
}

// Index an example for Swagger
//
//  @Summary      Summary
//  @Description  Description
//  @Tags         example
//  @Accept       json
//  @Success      200
//  @Failure      400
//  @Router       /swagger [get]
func (r *SwaggerController) Index(ctx http.Context) {
	ctx.Response().Success().Json(http.Json{
		"code": http.StatusOK,
	})
}

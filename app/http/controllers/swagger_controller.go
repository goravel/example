package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	"github.com/swaggo/http-swagger/v2"
	"sync"

	_ "goravel/docs"
)

/*********************************
1. Install swag
document: https://github.com/swaggo/http-swagger

go install github.com/swaggo/swag/cmd/swag@latest

2. Install http-swagger
go get -u github.com/swaggo/http-swagger

3. Optimize the document of endpoint: `app/http/controllers/swagger_controller.go`

4. Add route to `/route/web.go`

5. Init document
swag init

6. Run Server
air

7. Visit: http://localhost:3000/swagger/
 ********************************/

type SwaggerController struct {
	// Dependent services
}

func NewSwaggerController() *SwaggerController {
	return &SwaggerController{
		// Inject services
	}
}

// Index an example for Swagger
//
//	@Summary      Summary
//	@Description  Description
//	@Tags         example
//	@Accept       json
//	@Success      200
//	@Failure      400
//	@Router       /swagger [get]
//
// @Summary swagger index
// @Description swagger index
// @Tags Swagger
// @Router /swaggers [get]
func (r *SwaggerController) Index(ctx http.Context) http.Response {
	handler := httpSwagger.Handler()
	handler(ctx.Response().Writer(), ctx.Request().Origin())

	return nil
}

var (
	SwaggerControllerSingleton	*SwaggerController
	swaggerControllerOnce		sync.Once
)

func (r *SwaggerController) Singleton() *SwaggerController {

	swaggerControllerOnce.Do(func() {
		SwaggerControllerSingleton = NewSwaggerController()
	})

	return SwaggerControllerSingleton
}

// Routes Swagger routes.
// Example Usage:
// @api|web.go: controllers.SwaggerControllerSingleton.Routes(nil)
func (r *SwaggerController) Routes(baseRouter route.Router) {
	r.Singleton()
	var SwaggerRouter = baseRouter
	if SwaggerRouter == nil {
		SwaggerRouter = facades.Route()
	}
	SwaggerRouter.
		Get(
			"/swagger",

			SwaggerControllerSingleton.
				Index)
	SwaggerRouter.
		Get(
			"/swaggers",

			SwaggerControllerSingleton.
				Index)

}

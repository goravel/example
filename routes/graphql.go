package routes

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/graph"
	"goravel/graph/generated"
)

func Graphql() {
	facades.Route().Middleware(graphMiddleware).Any("/graphql", emptyController)
	facades.Route().Middleware(playgroundMiddleware).Get("/graphiql", emptyController)
}

type graphMiddlewareHandler struct{}

func (g *graphMiddlewareHandler) Signature() string {
	return "graphql"
}

func (g *graphMiddlewareHandler) Handle(ctx http.Context) {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
	h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
}

type playgroundMiddlewareHandler struct{}

func (p *playgroundMiddlewareHandler) Signature() string {
	return "graphql-playground"
}

func (p *playgroundMiddlewareHandler) Handle(ctx http.Context) {
	h := playground.Handler("GraphQL", "/graphql")
	h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
}

var (
	graphMiddleware      = &graphMiddlewareHandler{}
	playgroundMiddleware = &playgroundMiddlewareHandler{}
)

func emptyController(ctx http.Context) http.Response {
	return nil
}

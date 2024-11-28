package routes

import (
	"goravel/graph"
	"goravel/graph/generated"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func Graphql() {
	facades.Route().Middleware(graphMiddleware).Any("/graphql", emptyController)
	facades.Route().Middleware(playgroundMiddleware).Get("/graphiql", emptyController)
}

func graphMiddleware(ctx http.Context) {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
	h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
}

func playgroundMiddleware(ctx http.Context) {
	h := playground.Handler("GraphQL", "/graphql")
	h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
}

func emptyController(ctx http.Context) http.Response {
	return nil
}

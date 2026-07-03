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
	facades.Route().Middleware(&GraphqlMiddleware{}).Any("/graphql", emptyController)
	facades.Route().Middleware(&PlaygroundMiddleware{}).Get("/graphiql", emptyController)
}

type GraphqlMiddleware struct{}

func (g *GraphqlMiddleware) Handle(ctx http.Context) {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
	h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
}

func (g *GraphqlMiddleware) Signature() string {
	return "graphql"
}

type PlaygroundMiddleware struct{}

func (p *PlaygroundMiddleware) Handle(ctx http.Context) {
	h := playground.Handler("GraphQL", "/graphql")
	h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
}

func (p *PlaygroundMiddleware) Signature() string {
	return "playground"
}

func emptyController(ctx http.Context) http.Response {
	return nil
}

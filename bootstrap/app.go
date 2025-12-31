package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/http"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation"
	httpmiddleware "github.com/goravel/framework/http/middleware"
	"github.com/goravel/framework/session/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"goravel/app/events"
	"goravel/app/facades"
	"goravel/app/grpc/interceptors"
	"goravel/app/jobs"
	"goravel/app/listeners"
	"goravel/app/rules"
	"goravel/config"
	"goravel/database/seeders"
	"goravel/routes"
)

func Boot() {
	foundation.Setup().
		WithFilters(Filters()).
		WithCommands(Commands()).
		WithMigrations(Migrations()).
		WithSeeders([]seeder.Seeder{
			&seeders.DatabaseSeeder{},
		}).
		WithRouting([]func(){
			routes.Web,
			routes.Grpc,
			routes.Graphql,
		}).
		WithEvents(map[event.Event][]event.Listener{
			events.NewOrderShipped(): {
				listeners.NewSendShipmentNotification(),
			},
			events.NewOrderCanceled(): {
				listeners.NewSendShipmentNotification(),
			},
		}).
		WithJobs([]queue.Job{
			&jobs.Test{},
			&jobs.TestErr{},
		}).
		WithRules([]validation.Rule{
			&rules.Exists{},
			&rules.NotExists{},
		}).
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(httpmiddleware.Throttle("global"), middleware.StartSession()).Recover(func(ctx http.Context, err any) {
				facades.Log().Error(err)
				_ = ctx.Response().String(contractshttp.StatusInternalServerError, "recover").Abort()
			})
		}).
		WithGrpcServerInterceptors([]grpc.UnaryServerInterceptor{
			interceptors.OpentracingServer,
		}).
		WithGrpcClientInterceptors(map[string][]grpc.UnaryClientInterceptor{
			"default": {
				interceptors.OpentracingClient,
			},
		}).
		WithGrpcServerStatsHandlers([]stats.Handler{}).
		WithGrpcClientStatsHandlers(map[string][]stats.Handler{}).
		WithPaths(func(paths configuration.Paths) {
			paths.App("app")
		}).
		WithProviders(Providers()).
		WithSchedule([]schedule.Event{
			facades.Schedule().Call(func() {}),
		}).
		WithConfig(config.Boot).
		Run()
}

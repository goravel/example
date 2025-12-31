package bootstrap

import (
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/http"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/limit"
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

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(Filters()).
		WithCommands(Commands()).
		WithMigrations(Migrations()).
		WithSeeders([]seeder.Seeder{
			&seeders.DatabaseSeeder{},
		}).
		WithRouting([]func(){
			routes.Web,
			routes.Api,
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
		WithSchedule(func() []schedule.Event {
			return []schedule.Event{
				facades.Schedule().Call(func() {
					facades.Log().Info("Scheduled Task Executed")
				}),
			}
		}).
		WithCallback(func() {
			facades.RateLimiter().For("global", func(ctx contractshttp.Context) contractshttp.Limit {
				return limit.PerMinute(1000)
			})
			facades.RateLimiter().ForWithLimits("ip", func(ctx contractshttp.Context) []contractshttp.Limit {
				return []contractshttp.Limit{
					limit.PerDay(1000),
					limit.PerMinute(2).By(ctx.Request().Ip()),
				}
			})
			facades.Auth().Extend("another-jwt", auth.NewJwtGuard)
			facades.Auth().Provider("another-orm", auth.NewOrmUserProvider)
		}).
		WithConfig(config.Boot).
		Start()
}

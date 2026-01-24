package bootstrap

import (
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/event"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/http"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/limit"
	httpmiddleware "github.com/goravel/framework/http/middleware"
	"github.com/goravel/framework/session/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"goravel/app/events"
	"goravel/app/facades"
	"goravel/app/grpc/interceptors"
	"goravel/app/listeners"
	"goravel/app/models"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).
		WithSeeders(Seeders).
		WithRouting(func() {
			routes.Web()
			routes.Api()
			routes.Grpc()
			routes.Graphql()
		}).
		WithEvents(func() map[event.Event][]event.Listener {
			return map[event.Event][]event.Listener{
				events.NewOrderShipped(): {
					listeners.NewSendShipmentNotification(),
				},
				events.NewOrderCanceled(): {
					listeners.NewSendShipmentNotification(),
				},
			}
		}).
		WithJobs(Jobs).
		WithRules(Rules).
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				httpmiddleware.Throttle("global"),
				middleware.StartSession(),
			).Recover(func(ctx http.Context, err any) {
				facades.Log().Error(err)
				_ = ctx.Response().String(contractshttp.StatusInternalServerError, "recover").Abort()
			})
		}).
		WithGrpcServerInterceptors(func() []grpc.UnaryServerInterceptor {
			return []grpc.UnaryServerInterceptor{
				interceptors.TestServer,
			}
		}).
		WithGrpcClientInterceptors(func() map[string][]grpc.UnaryClientInterceptor {
			return map[string][]grpc.UnaryClientInterceptor{
				"default": {
					interceptors.TestClient,
				},
			}
		}).
		WithGrpcServerStatsHandlers(func() []stats.Handler {
			return []stats.Handler{}
		}).
		WithGrpcClientStatsHandlers(func() map[string][]stats.Handler {
			return map[string][]stats.Handler{}
		}).
		WithPaths(func(paths configuration.Paths) {
			paths.App("app")
		}).
		WithProviders(Providers).
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
			facades.Schema().Extend(schema.Extension{
				Models: []any{models.User{}},
			})
		}).
		WithConfig(config.Boot).
		Start()
}

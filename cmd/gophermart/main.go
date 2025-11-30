package main

import (
	"context"

	"github.com/SZabrodskii/gophermart-stas/internal/accrual"
	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/controllers"
	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/server"
	"github.com/SZabrodskii/gophermart-stas/internal/services"
	"github.com/SZabrodskii/gophermart-stas/internal/workers"
	"github.com/SZabrodskii/gophermart-stas/pkg/logger"

	"go.uber.org/fx"
)

func main() {
	fx.New(createApp()).Run()
}

func createApp() fx.Option {
	return fx.Options(
		logger.Module,
		services.Module,

		fx.Provide(
			config.New,
			database.New,
			accrual.ProvideClient,
			workers.NewAccrualWorker,
			server.NewGinEngine,
		),

		provideControllers(),

		fx.Invoke(func(engine *server.GinEngine, worker workers.AccrualWorkerI, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go worker.Start(ctx)
					go engine.Start(ctx)
					return nil
				},
			})
		}),
	)
}

func provideControllers() fx.Option {
	return fx.Provide(
		fx.Annotate(controllers.NewAuthControllerGin, fx.As(new(server.AuthControllerI))),
		fx.Annotate(controllers.NewOrderControllerGin, fx.As(new(server.OrderControllerI))),
		fx.Annotate(controllers.NewBalanceControllerGin, fx.As(new(server.BalanceControllerI))),
	)
}

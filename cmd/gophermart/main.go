package main

import (
	"github.com/SZabrodskii/gophermart-stas/internal/accrual"
	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/SZabrodskii/gophermart-stas/internal/controllers"
	"github.com/SZabrodskii/gophermart-stas/internal/database"
	"github.com/SZabrodskii/gophermart-stas/internal/server"
	"github.com/SZabrodskii/gophermart-stas/internal/services"
	"github.com/SZabrodskii/gophermart-stas/pkg/logger"

	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

func main() {
	fx.New(createApp()).Run()
}

func createApp() fx.Option {
	return fx.Options(
		logger.ZapModule,
		logger.HttpbaraLoggerModule,
		services.Module,

		fx.Provide(
			config.New,
			database.New,
			accrual.ProvideClient,
		),

		provideControllers(),
		server.ProvideHTTPModule("8080"),

		fx.Invoke(func(engine httpbara.Engine) {
		}),
	)
}

func provideControllers() fx.Option {
	return fx.Provide(
		controllers.NewAuthController,
		controllers.NewOrderController,
		controllers.NewBalanceController,
		controllers.NewJWTMiddleware,
	)
}

package services

import (
	"github.com/SZabrodskii/gophermart-stas/internal/domain"
	"go.uber.org/fx"
)

var Module = fx.Module("services",
	fx.Provide(
		fx.Annotate(
			NewOrderService,
			fx.As(new(domain.OrderServiceI)),
		),
		fx.Annotate(
			NewBalanceService,
			fx.As(new(domain.BalanceServiceI)),
		),
		fx.Annotate(
			NewUserService,
			fx.As(new(domain.UserServiceI)),
		),
	),
)

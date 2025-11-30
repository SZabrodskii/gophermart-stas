package server

import (
	"context"
	"strings"

	"github.com/SZabrodskii/gophermart-stas/internal/config"
	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
)

type AsHandlerOut struct {
	fx.Out

	Handler *httpbara.Handler `group:"handlers"`
}

func AsHandler(handler any) (AsHandlerOut, error) {
	h, err := httpbara.AsHandler(handler)
	if err != nil {
		return AsHandlerOut{}, err
	}
	return AsHandlerOut{Handler: h}, nil
}

type HTTPServerParams struct {
	Port string
}

type httpServerIn struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    httpbara.Logger
	Handlers  []*httpbara.Handler `group:"handlers"`
	Params    HTTPServerParams
}

func NewHTTPServer(in httpServerIn) (httpbara.Engine, error) {
	engine, err := httpbara.New(in.Handlers,
		httpbara.WithLogger(in.Logger),
	)
	if err != nil {
		return nil, err
	}

	in.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			in.Logger.Info("Starting HTTP server", "port", in.Params.Port)
			go func() {
				if err := engine.Run(":" + in.Params.Port); err != nil {
					in.Logger.Error("Server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			in.Logger.Info("Stopping HTTP server")
			return nil
		},
	})

	return engine, nil
}

func NewHTTPServerParams(cfg *config.Config) HTTPServerParams {
	port := cfg.RunAddress
	port = strings.TrimPrefix(port, ":")
	return HTTPServerParams{Port: port}
}

func ProvideHTTPModule() fx.Option {
	return fx.Options(
		fx.Provide(
			NewHTTPServerParams,
			NewHTTPServer,
		),
	)
}

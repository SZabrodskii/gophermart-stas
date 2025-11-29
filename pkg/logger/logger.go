package logger

import (
	"context"

	"github.com/gopybara/httpbara"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ZapLogger struct {
	log *zap.Logger
}

func NewZapLogger(log *zap.Logger) httpbara.Logger {
	return &ZapLogger{log: log}
}

func (l *ZapLogger) Info(message string, args ...any) {
	l.log.Info(message, l.mapFields(args...)...)
}

func (l *ZapLogger) Debug(message string, args ...any) {
	l.log.Debug(message, l.mapFields(args...)...)
}

func (l *ZapLogger) Error(message string, args ...any) {
	l.log.Error(message, l.mapFields(args...)...)
}

func (l *ZapLogger) Panic(message string, args ...any) {
	l.log.Panic(message, l.mapFields(args...)...)
}

func (l *ZapLogger) Warn(message string, args ...any) {
	l.log.Warn(message, l.mapFields(args...)...)
}

func (l *ZapLogger) mapFields(fields ...any) []zap.Field {
	expectingKey := true
	result := make([]zap.Field, 0)
	key := ""

	for i := 0; i < len(fields); i++ {
		switch field := fields[i].(type) {
		case zap.Field:
			result = append(result, field)
		case error:
			result = append(result, zap.Error(field))
		default:
			if expectingKey {
				if strKey, ok := field.(string); ok {
					key = strKey
				} else {
					key = ""
				}
			} else {
				var zapField zap.Field

				switch field.(type) {
				case string:
					zapField = zap.String(key, field.(string))
				case int:
					zapField = zap.Int(key, field.(int))
				case int64:
					zapField = zap.Int64(key, field.(int64))
				case uint:
					zapField = zap.Uint32(key, uint32(field.(uint)))
				case float64:
					zapField = zap.Float64(key, field.(float64))
				default:
					zapField = zap.Any(key, field)
				}

				result = append(result, zapField)
				key = ""
			}

			expectingKey = !expectingKey
		}
	}

	return result
}

var ZapModule = fx.Provide(NewZapInstance)

var HttpbaraLoggerModule = fx.Provide(NewZapLogger)

func NewZapInstance(lc fx.Lifecycle) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			_ = l.Sync()
			return nil
		},
	})
	return l, nil
}

var Module = ZapModule

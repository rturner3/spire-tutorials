package main

import (
	"context"
	"delegatedidentity/internal/caller"
	"delegatedidentity/internal/caller/periodic"
	"delegatedidentity/internal/common/configprovider"
	"delegatedidentity/internal/common/logger"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			configprovider.NewYAML,
			logger.NewDevelopment,
		),
		caller.Module,
		fx.Invoke(runClient),
		fx.WithLogger(
			func(logger *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{
					Logger: logger,
				}
			},
		),
	).Run()
}

func runClient(lc fx.Lifecycle, runner periodic.Runner) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			runner.Run()
			return nil
		},
		OnStop: func(context.Context) error {
			runner.Close()
			return nil
		},
	})
}

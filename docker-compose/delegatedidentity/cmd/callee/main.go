package main

import (
	"context"
	"delegatedidentity/internal/callee"
	"delegatedidentity/internal/callee/api/hello"
	"delegatedidentity/internal/common/configprovider"
	"delegatedidentity/internal/common/logger"
	"delegatedidentity/internal/common/spiffe"

	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	fx.New(
		fx.Provide(
			logger.NewDevelopment,
			configprovider.NewYAML,
			newX509SVIDSource,
		),
		hello.Module,
		callee.Module,
		fx.Invoke(run),
		fx.WithLogger(
			func(logger *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{
					Logger: logger,
				}
			},
		),
	).Run()
}

func newX509SVIDSource(lc fx.Lifecycle, c callee.Config, logger *zap.Logger) (x509svid.Source, x509bundle.Source, error) {
	x509Source, err := spiffe.NewX509SVIDSource(c.SPIRE.AgentAddr, logger.Sugar())
	if err != nil {
		return nil, nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return x509Source.Close()
		},
	})

	return x509Source, x509Source, nil
}

func run(lc fx.Lifecycle, config callee.Config, logger *zap.Logger, server *grpc.Server, helloServer *hello.Handler) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go callee.RunServer(config, logger, server, helloServer)
			return nil
		},
	})
}

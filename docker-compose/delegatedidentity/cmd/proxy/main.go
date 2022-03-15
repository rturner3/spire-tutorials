package main

import (
	"context"
	"delegatedidentity/internal/common/configprovider"
	"delegatedidentity/internal/common/logger"
	"delegatedidentity/internal/proxy"
	"delegatedidentity/internal/proxy/attestor"
	"delegatedidentity/internal/proxy/services/hello"
	"delegatedidentity/internal/proxy/spire"

	delegatedidentityv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/agent/delegatedidentity/v1"
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
			newDelegatedIdentityClient,
			attestor.New,
		),
		spire.Module,
		hello.Module,
		proxy.Module,
		fx.Invoke(runServer),
		fx.WithLogger(
			func(logger *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{
					Logger: logger,
				}
			},
		),
	).Run()
}

func runServer(lc fx.Lifecycle, x509BundleCache spire.X509BundleCache, config proxy.Config, server *grpc.Server, cc hello.ConnCache, logger *zap.Logger) {
	runCtx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			x509BundleCache.Init(runCtx)
			go proxy.RunProxy(config, server, logger)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			cc.Close()
			cancel()
			return nil
		},
	})
}

func newDelegatedIdentityClient(cfg proxy.Config) (delegatedidentityv1.DelegatedIdentityClient, error) {
	return spire.NewDelegatedIdentityClient(cfg.SPIRE.AgentSocketPath)
}

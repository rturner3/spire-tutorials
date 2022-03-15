package hello

import (
	"context"
	"delegatedidentity/internal/caller"
	"delegatedidentity/proto/hello"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ClientParams struct {
	fx.In

	Config caller.Config
	Logger *zap.Logger
}

func NewClient(p ClientParams) (hello.HelloClient, error) {
	ctx := context.Background()
	addr := p.Config.Outbounds.Proxy
	p.Logger.Debug("Dialing Hello server", zap.String("address", addr))
	cc, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return hello.NewHelloClient(cc), nil
}

package callee

import (
	"context"
	"net"

	"delegatedidentity/internal/callee/api/hello"
	hellopb "delegatedidentity/proto/hello"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewServer(lc fx.Lifecycle, creds credentials.TransportCredentials) *grpc.Server {
	s := grpc.NewServer(grpc.Creds(creds))
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			s.Stop()
			return nil
		},
	})

	return s
}

func RunServer(config Config, logger *zap.Logger, server *grpc.Server, helloServer *hello.Handler) error {
	var err error
	endpoint := net.JoinHostPort(config.Endpoints.GRPC.Address, config.Endpoints.GRPC.Port)
	tcpListener, err := net.Listen("tcp", endpoint)
	if err != nil {
		return err
	}

	hellopb.RegisterHelloServer(server, helloServer)
	return server.Serve(tcpListener)
}

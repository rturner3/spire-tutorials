package hello

import (
	"context"
	hellopb "delegatedidentity/proto/hello"

	"go.uber.org/zap"
)

type proxyServer struct {
	hellopb.HelloServer

	cc     ConnCache
	logger *zap.Logger
}

func NewProxyServer(cc ConnCache, logger *zap.Logger) hellopb.HelloServer {
	return &proxyServer{
		cc:     cc,
		logger: logger,
	}
}

func (p *proxyServer) SayHello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloReply, error) {
	cc, err := p.cc.Get(ctx)
	if err != nil {
		return nil, err
	}

	client := hellopb.NewHelloClient(cc)
	p.logger.Debug("Forwarding SayHello request to callee")
	return client.SayHello(ctx, req)
}

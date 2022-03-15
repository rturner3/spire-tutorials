package caller

import (
	"context"
	"time"
	"delegatedidentity/internal/caller/periodic"
	"delegatedidentity/proto/hello"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	helloTimeout = 30 * time.Second
	pollPeriod   = 5 * time.Second
)

type RunnerParams struct {
	fx.In

	Config Config
	Logger *zap.Logger
}

func NewRunner(p RunnerParams) periodic.Runner {
	return periodic.NewRunner(newTask(p.Config.ServiceName, p.Config.Outbounds.Proxy, p.Logger), pollPeriod)
}

type task struct {
	logger       *zap.Logger
	name         string
	outboundAddr string
}

func newTask(name, outboundAddr string, logger *zap.Logger) periodic.Task {
	return &task{
		logger:       logger,
		name:         name,
		outboundAddr: outboundAddr,
	}
}

func (t *task) Exec(ctx context.Context) {
	addr := t.outboundAddr
	t.logger.Debug("Dialing Hello server", zap.String("address", addr))
	cc, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.logger.Error("Dial failed", zap.Error(err))
	}

	defer cc.Close()
	client := hello.NewHelloClient(cc)

	reqCtx, cancel := context.WithTimeout(ctx, helloTimeout)
	defer cancel()
	req := &hello.HelloRequest{
		Name: t.name,
	}

	resp, err := client.SayHello(reqCtx, req)
	if err != nil {
		t.logger.Error("Hello request failed", zap.Error(err))
		return
	}

	t.logger.Info("Hello request succeeded", zap.String("resp_message", resp.Message))
}

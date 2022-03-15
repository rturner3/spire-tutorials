package hello

import (
	"context"
	"fmt"

	hellopb "delegatedidentity/proto/hello"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	Module = fx.Provide(
		NewHandler,
	)
)

type Handler struct {
	hellopb.UnsafeHelloServer

	logger *zap.Logger
}

func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) SayHello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloReply, error) {
	callerSpiffeID, ok := grpccredentials.PeerIDFromContext(ctx)
	logger := h.logger.With(zap.String("req_name", req.Name))
	if ok {
		logger = logger.With(zap.String("caller_spiffe_id", callerSpiffeID.String()))
	}

	logger.Info("Received SayHello request")
	return &hellopb.HelloReply{
		Message: fmt.Sprintf("Hello, %s", req.Name),
	}, nil
}

package spiffe

import (
	"context"

	"github.com/spiffe/go-spiffe/v2/logger"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func NewX509SVIDSource(addr string, logger logger.Logger) (*workloadapi.X509Source, error) {
	ctx := context.Background()
	s, err := workloadapi.NewX509Source(ctx,
		workloadapi.WithClientOptions(
			workloadapi.WithAddr(addr),
			workloadapi.WithLogger(logger),
		),
	)

	if err != nil {
		return nil, err
	}

	return s, nil
}

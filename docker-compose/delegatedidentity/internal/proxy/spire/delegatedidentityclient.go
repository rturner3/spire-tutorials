package spire

import (
	"context"
	"time"

	delegatedidentityv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/agent/delegatedidentity/v1"
	"google.golang.org/grpc"
)

const (
	dialTimeout = 30 * time.Second
)

func NewDelegatedIdentityClient(spireAgentUnixSocketPath string) (delegatedidentityv1.DelegatedIdentityClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, spireAgentUnixSocketPath, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return delegatedidentityv1.NewDelegatedIdentityClient(conn), nil
}

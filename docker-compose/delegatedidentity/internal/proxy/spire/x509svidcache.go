package spire

import (
	"context"
	"crypto/x509"
	"delegatedidentity/internal/proxy/attestor"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	delegatedidentityv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/agent/delegatedidentity/v1"
	"github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"go.uber.org/zap"
)

var (
	noSVIDError = errors.New("no SVID for caller")
)

// X509SVIDCache is a read-through cache that fetches SVIDs from SPIRE Agent over the SubscribeToX509SVIDs method
// on a cache miss.
type X509SVIDCache interface {
	X509SVID(context.Context, []*attestor.Selector) (*delegatedidentityv1.X509SVIDWithKey, error)
}

type x509SVIDCache struct {
	client delegatedidentityv1.DelegatedIdentityClient
	logger *zap.Logger
	mu     sync.RWMutex
	svids  map[string][]*delegatedidentityv1.X509SVIDWithKey
}

func NewX509SVIDCache(client delegatedidentityv1.DelegatedIdentityClient, logger *zap.Logger) X509SVIDCache {
	return &x509SVIDCache{
		client: client,
		logger: logger,
		svids:  make(map[string][]*delegatedidentityv1.X509SVIDWithKey),
	}
}

func (c *x509SVIDCache) X509SVID(ctx context.Context, selectors []*attestor.Selector) (*delegatedidentityv1.X509SVIDWithKey, error) {
	key := selectorsKey(selectors)
	if svids, ok := c.svidsFromCache(key); ok {
		if len(svids) == 0 {
			return nil, noSVIDError
		}

		svidDER := svids[0].X509Svid.CertChain[0]
		svid, err := x509.ParseCertificate(svidDER)
		if err != nil {
			return nil, fmt.Errorf("certificate is not valid ASN.1 DER: %s", err)
		}

		if svid.NotAfter.After(time.Now()) {
			return svids[0], nil
		}

		// The cert has expired, which may indicate that we might have lost the stream to SPIRE Agent
		// Explicitly remove the SVID from the cache so that we can retry getting it from SPIRE
		c.removeSVIDs(key)
	}

	retCh := make(chan []*delegatedidentityv1.X509SVIDWithKey, 1)
	errCh := make(chan error, 1)
	go c.subscribeToSelectors(context.Background(), selectors, key, retCh, errCh)
	select {
	case <-ctx.Done():
		return nil, errors.New("timed out waiting for identity")
	case err := <-errCh:
		return nil, err
	case svids := <-retCh:
		if len(svids) > 0 {
			return svids[0], nil
		}

		return nil, noSVIDError
	}
}

func (c *x509SVIDCache) svidsFromCache(key string) ([]*delegatedidentityv1.X509SVIDWithKey, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	svids, ok := c.svids[key]
	return svids, ok
}

func (c *x509SVIDCache) subscribeToSelectors(ctx context.Context, selectors []*attestor.Selector, key string, retCh chan<- []*delegatedidentityv1.X509SVIDWithKey, errCh chan<- error) {
	req := &delegatedidentityv1.SubscribeToX509SVIDsRequest{
		Selectors: toProtoSelectors(selectors),
	}

	stream, err := c.client.SubscribeToX509SVIDs(ctx, req)
	if err != nil {
		c.logger.Warn("Failed to initialize stream to SPIRE Agent", zap.Error(err))
		errCh <- err
		return
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			c.logger.Error("Received error from SPIRE Agent stream", zap.Error(err))
			errCh <- err
			return
		}

		svids := resp.GetX509Svids()
		c.setSVIDs(key, svids)
		retCh <- svids
		if len(svids) > 0 {
			spiffeID := fmt.Sprintf("spiffe://%s%s", svids[0].X509Svid.Id.TrustDomain, svids[0].X509Svid.Id.Path)
			c.logger.Debug("Received SVIDs for workload",
				zap.String("selectors", key),
				zap.String("spiffe_id", spiffeID),
			)
		}
	}
}

func (c *x509SVIDCache) setSVIDs(key string, svids []*delegatedidentityv1.X509SVIDWithKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.svids[key] = svids
}

func (c *x509SVIDCache) removeSVIDs(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.svids, key)
}

func selectorsKey(selectors []*attestor.Selector) string {
	selectorStrings := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		selectorString := fmt.Sprintf("%s:%s", selector.Type, selector.Value)
		selectorStrings = append(selectorStrings, selectorString)
	}

	sort.Strings(selectorStrings)
	return strings.Join(selectorStrings, ",")
}

func toProtoSelectors(selectors []*attestor.Selector) []*types.Selector {
	protoSelectors := make([]*types.Selector, 0, len(selectors))
	for _, selector := range selectors {
		protoSelector := &types.Selector{
			Type:  selector.Type,
			Value: selector.Value,
		}

		protoSelectors = append(protoSelectors, protoSelector)
	}

	return protoSelectors
}

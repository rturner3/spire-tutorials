package spire

import (
	"context"
	"fmt"
	"io"
	"sync"

	delegatedidentityv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/agent/delegatedidentity/v1"
	"go.uber.org/zap"
)

type X509BundleCache interface {
	Init(ctx context.Context)
	X509Bundle(ctx context.Context, trustDomain string) ([]byte, error)
}

type x509BundleCache struct {
	bundles               map[string][]byte
	client                delegatedidentityv1.DelegatedIdentityClient
	firstBundleReceivedCh chan struct{}
	receivedFirstBundle   bool
	logger                *zap.Logger
	mu                    sync.RWMutex
}

func NewX509BundleCache(client delegatedidentityv1.DelegatedIdentityClient, logger *zap.Logger) X509BundleCache {
	return &x509BundleCache{
		bundles:               make(map[string][]byte),
		client:                client,
		firstBundleReceivedCh: make(chan struct{}),
		logger:                logger,
	}
}

func (c *x509BundleCache) Init(ctx context.Context) {
	go func() {
		req := &delegatedidentityv1.SubscribeToX509BundlesRequest{}
		stream, err := c.client.SubscribeToX509Bundles(ctx, req)
		if err != nil {
			c.logger.Error("Failed to open X.509 bundles stream to SPIRE Agent", zap.Error(err))
			return
		}

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}

			if err != nil {
				c.logger.Error("Failed to get a response over bundle stream", zap.Error(err))
				return
			}

			trustDomains := make([]string, len(resp.CaCertificates))
			i := 0
			for k := range resp.CaCertificates {
				trustDomains[i] = k
				i++
			}

			c.logger.Debug("Received response from SubscribeToX509Bundles", zap.Strings("trust_domains", trustDomains))
			c.setBundles(resp.CaCertificates)
			c.firstBundleReceivedCh <- struct{}{}
		}
	}()
}

func (c *x509BundleCache) X509Bundle(ctx context.Context, trustDomain string) ([]byte, error) {
	c.awaitFirstBundleResponse(ctx)
	bundle, ok := c.x509BundleFromCache(trustDomain)
	if !ok {
		return nil, fmt.Errorf("missing trust bundle for trust domain %s", trustDomain)
	}

	return bundle, nil
}

func (c *x509BundleCache) setBundles(bundles map[string][]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bundles = bundles
	c.receivedFirstBundle = true
}

func (c *x509BundleCache) awaitFirstBundleResponse(ctx context.Context) error {
	if c.firstBundleReceived() {
		return nil
	}

	select {
	case <-c.firstBundleReceivedCh:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for first bundle: %w", ctx.Err())
	}
}

func (c *x509BundleCache) firstBundleReceived() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.receivedFirstBundle
}

func (c *x509BundleCache) x509BundleFromCache(trustDomain string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	bundle, ok := c.bundles[trustDomain]
	return bundle, ok
}

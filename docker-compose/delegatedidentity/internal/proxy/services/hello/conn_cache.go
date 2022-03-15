package hello

import (
	"context"
	"crypto/x509"
	"delegatedidentity/internal/proxy"
	"delegatedidentity/internal/proxy/attestor"
	"delegatedidentity/internal/proxy/spire"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/spire/pkg/common/peertracker"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type ConnCache interface {
	Get(context.Context) (*grpc.ClientConn, error)
	Close()
}

type clientConn struct {
	conn       *grpc.ClientConn
	expiration time.Time
	mu         sync.RWMutex
}

func newClientConn(conn *grpc.ClientConn, certNotAfter time.Time) *clientConn {
	return &clientConn{
		conn:       conn,
		expiration: certNotAfter.Add(-time.Minute),
	}
}

func (c *clientConn) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.Close()
	c.conn = nil
}

func (c *clientConn) Conn() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

type connCache struct {
	att             attestor.Interface
	backend         string
	clients         map[string]*clientConn
	mu              sync.RWMutex
	logger          *zap.Logger
	trustDomain     spiffeid.TrustDomain
	x509BundleCache spire.X509BundleCache
	x509SVIDCache   spire.X509SVIDCache
}

type ConnCacheParams struct {
	fx.In

	Attestor        attestor.Interface
	Config          proxy.Config
	Logger          *zap.Logger
	X509BundleCache spire.X509BundleCache
	X509SVIDCache   spire.X509SVIDCache
}

func NewConnCache(p ConnCacheParams) (ConnCache, error) {
	td, err := spiffeid.TrustDomainFromString(p.Config.SPIRE.TrustDomain)
	if err != nil {
		return nil, err
	}

	p.X509BundleCache.Init(context.Background())
	return &connCache{
		att:             p.Attestor,
		backend:         p.Config.Backend,
		clients:         make(map[string]*clientConn),
		logger:          p.Logger,
		trustDomain:     td,
		x509BundleCache: p.X509BundleCache,
		x509SVIDCache:   p.X509SVIDCache,
	}, nil
}

func (c *connCache) Get(ctx context.Context) (*grpc.ClientConn, error) {
	logger := c.logger
	watcher, ok := peertracker.WatcherFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "peer tracker watcher missing from context")
	}

	selectors, err := c.att.Attest(ctx, watcher.PID())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "could not verify existence of the original caller: %w", err)
	}

	if caller, ok := peer.FromContext(ctx); ok {
		logger = logger.With(zap.String("caller_addr", caller.Addr.String()))
	}

	logger = logger.With(zap.Int32("caller_pid", watcher.PID()))
	logger.Info("Received SayHello request")
	svid, err := c.x509SVIDCache.X509SVID(ctx, selectors)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "caller does not have an identity: %s", err.Error())
	}

	spiffeId, err := spiffeid.FromURI(&url.URL{
		Scheme: "spiffe",
		Host:   svid.X509Svid.Id.TrustDomain,
		Path:   svid.X509Svid.Id.Path,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "caller has malformed SPIFFE ID: %s", err.Error())
	}

	logger = logger.With(zap.String("caller_spiffe_id", spiffeId.String()))

	key := selectorsKey(selectors)
	if cc, ok := c.getConn(key); ok {
		if cc.expiration.After(time.Now()) {
			return cc.Conn(), nil
		}

		logger.Debug("Expiring existing connection")
		c.removeConn(key)
		cc.Close()
	}

	bundle, err := c.x509BundleCache.X509Bundle(ctx, c.trustDomain.IDString())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "no trust bundle for configured trust domain %s: %s", c.trustDomain, err.Error())
	}

	var x509SvidCerts []byte
	for _, cert := range svid.X509Svid.CertChain {
		x509SvidCerts = append(x509SvidCerts, cert...)
	}

	x509SvidSource, err := x509svid.ParseRaw(x509SvidCerts, svid.X509SvidKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert X.509-SVID to source: %s", err.Error())
	}

	x509BundleCerts, err := x509.ParseCertificates(bundle)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse X.509 bundle: %s", err.Error())
	}

	x509BundleSource := x509bundle.FromX509Authorities(c.trustDomain, x509BundleCerts)
	creds := grpccredentials.MTLSClientCredentials(x509SvidSource, x509BundleSource, tlsconfig.AuthorizeMemberOf(c.trustDomain))

	logger.Debug("Dialing HelloServer")
	conn, err := grpc.DialContext(ctx, c.backend,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to dial backend: %s", err.Error())
	}

	logger.Debug("Established connection to HelloServer")

	cc := newClientConn(conn, x509SvidSource.Certificates[0].NotAfter)
	c.setConn(key, cc)
	return cc.Conn(), nil
}

func (c *connCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, client := range c.clients {
		client.Close()
	}

	c.clients = nil
}

func (c *connCache) setConn(key string, cc *clientConn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clients[key] = cc
}

func (c *connCache) getConn(key string) (*clientConn, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	clientConn, ok := c.clients[key]
	return clientConn, ok
}

func (c *connCache) removeConn(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.clients, key)
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

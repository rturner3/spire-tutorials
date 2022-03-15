package callee

import (
	"net"
	"delegatedidentity/internal/common/spiffe"

	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

type credsWrapper struct {
	credentials.TransportCredentials
	logger *zap.Logger
}

func NewMTLSServerCredentials(svid x509svid.Source, bundle x509bundle.Source, logger *zap.Logger) credentials.TransportCredentials {
	return &credsWrapper{
		TransportCredentials: grpccredentials.MTLSServerCredentials(svid, bundle, tlsconfig.AuthorizeMemberOf(spiffe.TrustDomain)),
		logger:               logger,
	}
}

func (c *credsWrapper) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	cc, authInfo, err := c.TransportCredentials.ServerHandshake(conn)
	if err != nil {
		c.logger.Debug("Server TLS handshake failed", zap.Error(err))
	}

	return cc, authInfo, err
}

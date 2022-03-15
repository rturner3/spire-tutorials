package caller

import (
	"delegatedidentity/internal/common/spiffe"

	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"google.golang.org/grpc/credentials"
)

func NewMTLSClientCredentials(svid x509svid.Source, bundle x509bundle.Source) credentials.TransportCredentials {
	return grpccredentials.MTLSClientCredentials(svid, bundle, tlsconfig.AuthorizeMemberOf(spiffe.TrustDomain))
}

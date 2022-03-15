package spiffe

import "github.com/spiffe/go-spiffe/v2/spiffeid"

var (
	TrustDomain = spiffeid.RequireTrustDomainFromString("spiffe://example.org")
)

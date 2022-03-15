package spire

import "go.uber.org/fx"

var (
	Module = fx.Provide(
		NewX509BundleCache,
		NewX509SVIDCache,
	)
)

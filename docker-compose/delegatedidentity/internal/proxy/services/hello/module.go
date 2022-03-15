package hello

import "go.uber.org/fx"

var (
	Module = fx.Provide(
		NewConnCache,
		NewProxyServer,
	)
)

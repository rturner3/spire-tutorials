package callee

import (
	"go.uber.org/fx"
)

var (
	Module = fx.Provide(
		NewMTLSServerCredentials,
		NewConfig,
		NewServer,
	)
)

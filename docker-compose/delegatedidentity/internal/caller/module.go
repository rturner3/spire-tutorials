package caller

import "go.uber.org/fx"

var (
	Module = fx.Provide(
		NewConfig,
		NewRunner,
	)
)

package periodic

import "context"

type Task interface {
	Exec(context.Context)
}

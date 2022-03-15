package periodic

import (
	"context"
	"time"
)

type Runner interface {
	Run()
	Close()
}

type runner struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	d         time.Duration
	t         Task
}

func NewRunner(t Task, d time.Duration) Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &runner{
		ctx:       ctx,
		ctxCancel: cancel,
		d:         d,
		t:         t,
	}
}

func (r *runner) Run() {
	go func() {
		r.t.Exec(r.ctx)
		for {
			select {
			case <-time.After(r.d):
			case <-r.ctx.Done():
				return
			}
			r.t.Exec(r.ctx)
		}
	}()
}

func (r *runner) Close() {
	r.ctxCancel()
}

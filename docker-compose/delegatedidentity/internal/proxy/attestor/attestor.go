package attestor

import (
	"context"

	workloadattestorv1 "github.com/spiffe/spire-plugin-sdk/proto/spire/plugin/agent/workloadattestor/v1"
	configv1 "github.com/spiffe/spire-plugin-sdk/proto/spire/service/common/config/v1"
	"github.com/spiffe/spire/pkg/agent/plugin/workloadattestor/docker"
	"go.uber.org/zap"
)

type Selector struct {
	Type  string
	Value string
}

type Interface interface {
	Attest(context.Context, int32) ([]*Selector, error)
}

type attestor struct {
	dockerPlugin *docker.Plugin
	logger       *zap.Logger
}

func New(logger *zap.Logger) Interface {
	dockerPlugin := docker.New()
	dockerPlugin.Configure(context.Background(), &configv1.ConfigureRequest{})
	return &attestor{
		dockerPlugin: dockerPlugin,
		logger:       logger,
	}
}

func (a *attestor) Attest(ctx context.Context, pid int32) ([]*Selector, error) {
	var selectors []*Selector
	resp, err := a.dockerPlugin.Attest(ctx, &workloadattestorv1.AttestRequest{
		Pid: pid,
	})

	if err != nil {
		a.logger.Error("Failed to discover selectors", zap.Error(err))
		return nil, err
	}

	for _, sel := range resp.SelectorValues {
		selector := &Selector{
			Type:  "docker",
			Value: sel,
		}
		selectors = append(selectors, selector)
	}

	return selectors, nil
}

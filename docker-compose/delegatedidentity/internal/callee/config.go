package callee

import "go.uber.org/config"

type Config struct {
	Endpoints EndpointsConfig `yaml:"endpoints"`
	SPIRE     SPIREConfig     `yaml:"spire"`
}

type EndpointsConfig struct {
	GRPC GRPCConfig `yaml:"grpc"`
}

type GRPCConfig struct {
	Address string
	Port    string
}

type SPIREConfig struct {
	AgentAddr string `yaml:"agentAddr"`
}

func NewConfig(cp config.Provider) (Config, error) {
	var cc Config
	err := cp.Get(config.Root).Populate(&cc)
	return cc, err
}

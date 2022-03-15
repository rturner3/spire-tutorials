package proxy

import "go.uber.org/config"

type Config struct {
	Backend          string
	ListenSocketPath string      `yaml:"listenSocketPath"`
	SPIRE            SPIREConfig `yaml:"spire"`
}

type SPIREConfig struct {
	AgentSocketPath string `yaml:"agentSocketPath"`
	TrustDomain     string `yaml:"trustDomain"`
}

func NewConfig(cp config.Provider) (Config, error) {
	var cc Config
	err := cp.Get(config.Root).Populate(&cc)
	return cc, err
}

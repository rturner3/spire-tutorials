package caller

import "go.uber.org/config"

type Config struct {
	ServiceName string `yaml:"serviceName"`
	Outbounds   Outbounds
}

type Outbounds struct {
	Proxy string
}

func NewConfig(cp config.Provider) (Config, error) {
	var cc Config
	err := cp.Get(config.Root).Populate(&cc)
	return cc, err
}

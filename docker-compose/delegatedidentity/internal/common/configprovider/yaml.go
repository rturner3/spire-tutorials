package configprovider

import (
	"os"
	"path"

	"go.uber.org/config"
)

const (
	baseYamlFile    = "base.yaml"
	configDirEnvVar = "CONFIG_DIR"
)

func NewYAML() (config.Provider, error) {
	var configDir string
	var err error
	var ok bool
	configDir, ok = os.LookupEnv(configDirEnvVar)
	if !ok {
		configDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	baseConfigFile := path.Join(configDir, baseYamlFile)
	provider, err := config.NewYAML(config.File(baseConfigFile))
	if err != nil {
		return nil, err
	}

	return provider, nil
}

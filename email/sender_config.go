package email

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

func LoadSenderConfig(configFileYaml string, cfg *Sender) (err error) {
	var file *os.File
	if file, err = os.Open(configFileYaml); err != nil {
		return
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(cfg); err != nil {
		return
	}
	err = envconfig.Process("", cfg)
	return
}

func (o *EngineConfig) WriteSenderConfig(configFileYaml string) (err error) {
	var file *os.File
	if file, err = os.Create(configFileYaml); err != nil {
		return
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	if err = encoder.Encode(o); err != nil {
		return
	}
	err = encoder.Close()
	return
}

package email

import (
	"testing"
)

func TestEngineSend(t *testing.T) {
	err := sendByEngine()
	if err != nil {
		t.Error(err)
	}
}

func sendByEngine() (err error) {
	var config EngineConfig
	if err = EngineConfigFileYamlLoad("engine-config.yml", &config); err != nil {
		return
	}

	if err = config.Setup(); err != nil {
		return
	}

	var engine *Engine
	if engine, err = NewEngine(&config); err != nil {
		return
	}

	err = engine.Send(emailDataDefault())
	return
}

func emailDataDefault() *EmailData {

	return &EmailData{
		To:      []string{"eoeisler@gmail.com"},
		Name:    "TestEugen",
		Subject: "Berufung: Anmeldung Seminar",
		Url:     "www.reguel.de",
		Theme:   "default2",
		Markdown:
		`# Test
dsfs
sdfsd
## sdfsd
### sdf
fsdf
sdf
sdf
sd`,
	}
}

package formatter

import (
	"github.com/VertexC/log-formatter/formatter/general"
)

type Formatter interface {
	Init(filePath string, verbose bool)
	Format(msg string) map[string]interface{}
}

type Config struct {
	Type       string         `yaml:"type"`
	GeneralCfg general.Config `yaml:"general"`
}

func New(config Config) Formatter {
	switch config.Type {
	case "general":
		formatter := new(general.Formatter)
		formatter.SetConfig(config.GeneralCfg)
		return formatter
	}
	return nil
}

func Execute(config Config, inputCh chan interface{}, outputCh chan interface{}, filePath string, verbose bool) {
	formatter := New(config)
	formatter.Init(filePath, verbose)

	for {
		record := <-inputCh
		// make message field configurable
		message := record.(map[string]interface{})["message"].(string)
		// FIXME: here kv should be map[string]string

		kvMap := formatter.Format(message)
		if kvMap == nil {
			kvMap = map[string]interface{}{"sourceData_": record}
			continue
		} else {
			kvMap["sourceData_"] = record
		}
		outputCh <- kvMap
	}
}

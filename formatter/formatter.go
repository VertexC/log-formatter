package formatter

import (
	"github.com/VertexC/log-formatter/formatter/general"
	"log"
)

type Formatter interface {
	Init(logPath string, verbose bool)
	Format(msg string) map[string]interface{}
}

type Config struct {
	Type       string         `yaml:"type"`
	GeneralCfg general.Config `yaml:"general"`
}

func New(config Config, logPath string, verbose bool) Formatter {
	switch config.Type {
	case "general":
		formatter := new(general.Formatter)
		formatter.SetConfig(config.GeneralCfg)
		formatter.Init(logPath, verbose)
		return formatter
	case "":
		return nil
	default:
		log.Fatalf("Invalid Logger %s\n", config.Type)
	}
	return nil
}

func Execute(config Config, inputCh chan interface{}, outputCh chan interface{}, logPath string, verbose bool) {
	formatter := New(config, logPath, verbose)

	for {
		record := <-inputCh
		// make message field configurable
		message := record.(map[string]interface{})["message"].(string)
		// FIXME: strict kvMap here into map[string]string?
		var kvMap map[string]interface{}
		// FIXME: bad if inside loop
		if formatter == nil {
			outputCh <- record
		} else {
			kvMap = formatter.Format(message)
			if kvMap == nil {
				continue
			} else {
				kvMap["sourceData_"] = record
			}
			outputCh <- kvMap
		}
	}
}

package formatter

import (
	"github.com/VertexC/log-formatter/formatter/general"
	"log"
)

type Formatter interface {
	Init(logPath string, verbose bool)
	Format(msg string) map[string]interface{}
}

type Label struct {
	Key string `yaml: "key"`
	Val string `yaml: "val"`
}

type Config struct {
	Type       string         `yaml:"type"`
	Labels     []Label        `yaml:"labels"`
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

// TODO: maybe move to util
func Merge(a map[string]interface{}, b map[string]interface{}) {
	for k, v := range b {
		a[k] = v
	}
}

func Execute(config Config, inputCh chan interface{}, outputCh chan interface{}, logPath string, verbose bool) {
	formatter := New(config, logPath, verbose)
	labels := map[string]interface{}{}
	for _, label := range config.Labels {
		labels[label.Key] = label.Val
	}
	for {
		result := map[string]interface{}{}
		Merge(result, labels)

		record := <-inputCh
		// make message field configurable
		message := record.(map[string]interface{})["message"].(string)
		// FIXME: strict kvMap here into map[string]string?
		var kvMap map[string]interface{}
		// FIXME: bad if inside loop
		if formatter == nil {
			Merge(result, record.(map[string]interface{}))
			// log.Fatalln(labels)
			outputCh <- result
		} else {
			kvMap = formatter.Format(message)
			if kvMap == nil {
				continue
			} else {
				kvMap["sourceData_"] = record
			}
			Merge(result, kvMap)
			outputCh <- result
		}
	}
}

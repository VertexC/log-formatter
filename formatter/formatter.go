package formatter

import (
	"github.com/VertexC/log-formatter/formatter/general"
	"log"
	"regexp"
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
	Type          string         `yaml:"type"`
	Labels        []Label        `yaml:"labels"`
	GeneralCfg    general.Config `yaml:"general"`
	IncludeFields []string       `yaml:"include_fields"`
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

// TODO: Cache
func Filter(includeFields []string, record map[string]interface{}) map[string]interface{} {
	regList := []*regexp.Regexp{}
	for _, s := range includeFields {
		regList = append(regList, regexp.MustCompile(s))
	}

	result := map[string]interface{}{}
	for k, v := range record {
		for _, r := range regList {
			if r.MatchString(k) {
				result[k] = v
				break
			}
		}
	}
	return result
}

func Execute(config Config, inputCh chan map[string]interface{}, outputCh chan interface{}, logPath string, verbose bool) {
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
		message := record["message"].(string)
		// filter fields from source data
		record = Filter(config.IncludeFields, record)
		Merge(result, record)
		// FIXME: strict kvMap here into map[string]string?
		var kvMap map[string]interface{}
		// FIXME: bad if inside loop
		if formatter == nil {
			Merge(result, record)
			// log.Fatalln(labels)
			outputCh <- result
		} else {
			kvMap = formatter.Format(message)
			if kvMap == nil {
				continue
			}
			Merge(result, kvMap)
			outputCh <- result
		}
	}
}

package general

import (
	"fmt"
	"github.com/VertexC/log-formatter/util"
	"regexp"
)

type Label struct {
	Component string   `yaml:"component"`
	Regexprs  []string `yaml:"regexprs"`
}

type Config struct {
	Regex  string  `yaml:"components"`
	Labels []Label `yaml:"labels"`
}

type Formatter struct {
	config Config
	logger util.Logger
}

func (formatter *Formatter) SetConfig(config Config) {
	formatter.config = config
}

func (formatter *Formatter) Init(logFile string, verbose bool) {
	formatter.logger.Init(logFile, "General-Formatter", verbose)
}

func (formatter *Formatter) DiscardLog() {
	formatter.logger.DiscardAll()
}

func (formatter *Formatter) Format(msg string) map[string]interface{} {
	// FIXME: okay to allow panic happens and terminate the process?
	componentMap, err := reSubMatchMap(formatter.config.Regex, msg)
	if err != nil {
		formatter.logger.Error.Printf("Error occurs while get compponents: %s\n", err)
		return nil
	}

	kvResult := map[string]interface{}{}

	for _, label := range formatter.config.Labels {
		if component, ok := componentMap[label.Component]; !ok {
			formatter.logger.Error.Printf("Componet %s not found in matchMap %+v: %s\n", label.Component, componentMap, err)
		} else {
			for _, regex := range label.Regexprs {
				labelMap, err := reSubMatchMap(regex, component)
				if err != nil {
					formatter.logger.Warning.Printf("Error occurs while get labels: %s\n", err)
					continue
				}
				for key, val := range labelMap {
					if key[len(key)-1:] == "_" {
						continue
					}
					kvResult[key] = val
				}
			}
		}
	}

	for key, val := range componentMap {
		if key[len(key)-1:] == "_" {
			continue
		}
		kvResult[key] = val
	}

	return kvResult
}

func reSubMatchMap(reg string, str string) (map[string]string, error) {
	r := regexp.MustCompile(reg)
	match := r.FindStringSubmatch(str)
	groupNames := r.SubexpNames()
	if len(match) != len(groupNames) {
		return nil, fmt.Errorf("Failed to extract groups %s from %s with %s, match:%s", r.SubexpNames(), str, reg, match)
	}
	subMatchMap := map[string]string{}
	for i, name := range groupNames {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}
	return subMatchMap, nil
}

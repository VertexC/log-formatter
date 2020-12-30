package pipeline

import (
	"fmt"

	"github.com/VertexC/log-formatter/util"
)

type Label struct {
	Key string `yaml: "key"`
	Val string `yaml: "val"`
}

type Formatter interface {
	Format(map[string]interface{}) (map[string]interface{}, error)
}

type Factory = func(interface{}) (Formatter, error)

var registry = make(map[string]Factory)
var logger = util.NewLogger("PIPLINE")

func Register(name string, factory Factory) error {
	logger.Info.Printf("Registering formatter <%s>\n", name)
	if name == "" {
		return fmt.Errorf("Error registering formatter: name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("Error registering formatter '%v': factory cannot be empty", name)
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf("Error registering formatter '%v': already registered", name)
	}

	registry[name] = factory
	logger.Info.Printf("Successfully registered formatter <%s>\n", name)

	return nil
}

func NewFormatter(content interface{}) (Formatter, error) {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot convert given formatter config to mapStr")
	}
	for name, val := range contentMapStr {
		if factory, ok := registry[name]; ok {
			output, err := factory(val)
			return output, err
		}
	}
	return nil, fmt.Errorf("Failed to creat any output target")
}

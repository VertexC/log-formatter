package pipeline

import (
	"fmt"

	"github.com/VertexC/log-formatter/agent/pipeline/protocol"
	"github.com/VertexC/log-formatter/util"
)

type Label struct {
	Key string `yaml: "key"`
	Val string `yaml: "val"`
}

type Factory = func(interface{}) (protocol.Formatter, error)

var registry = make(map[string]Factory)

func Register(name string, factory Factory) error {
	util.UseLog(TAG).Info.Printf("Registering formatter <%s>\n", name)
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
	util.UseLog(TAG).Info.Printf("Successfully registered formatter <%s>\n", name)

	return nil
}

func NewFormatter(content interface{}) (protocol.Formatter, error) {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot convert given formatter config to mapStr")
	}
	for name, val := range contentMapStr {
		if factory, ok := registry[name]; ok {
			formatter, err := factory(val)
			return formatter, err
		} else if util.IsSoFile(name) {
			formatter, err := loadFormatterPlugin(name, val)
			return formatter, err
		}
	}
	return nil, fmt.Errorf("Failed to creat any formatter target")
}

func loadFormatterPlugin(url string, content interface{}) (protocol.Formatter, error) {
	p, err := util.LoadPlugin(url)
	if err != nil {
		return nil, fmt.Errorf("Could not load plugin from url %s: %s", url, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("could not find New function in %s: %s", url, err)
	}
	f, ok := newFunc.(Factory)
	if !ok {
		return nil, fmt.Errorf("`New` func in %s doesn't implement formatter interface", url)
	}
	instance, err := f(content)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

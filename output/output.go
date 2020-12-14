package output

import (
	"fmt"

	"github.com/VertexC/log-formatter/util"
)

type Output interface {
	Run()
}

type Factory = func(interface{}, chan map[string]interface{}) (Output, error)

var registry = make(map[string]Factory)
var logger = util.NewLogger("OUTPUT")

func Register(name string, factory Factory) error {
	logger.Info.Printf("Registering output <%s>\n", name)
	if name == "" {
		return fmt.Errorf("Error registering input: name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("Error registering output '%v': factory cannot be empty", name)
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf("Error registering output '%v': already registered", name)
	}

	registry[name] = factory
	logger.Info.Printf("Successfully registered output <%s>\n", name)

	return nil
}

func NewOutput(content interface{}, docCh chan map[string]interface{}) (Output, error) {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot convert given config to mapStr")
	}
	if len(contentMapStr) > 1 {
		return nil, fmt.Errorf("Cannot have multiple output targets.")
	}
	for target, val := range contentMapStr {
		if factory, ok := registry[target]; ok {
			output, err := factory(val, docCh)
			return output, err
		}
	}
	return nil, fmt.Errorf("Failed to creat any output target")
}

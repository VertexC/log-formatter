package input

import (
	"fmt"
	"github.com/VertexC/log-formatter/util"
)

type Input interface {
	// TODO: wrap inputCh and outputCh into contextChannl
	Run()
}

type Factory = func(interface{}, chan util.Doc) (Input, error)

var registry = make(map[string]Factory)
var logger = util.NewLogger("INPUT")

func Register(name string, factory Factory) error {
	logger.Info.Printf("Registering input <%s>\n", name)
	if name == "" {
		return fmt.Errorf("Error registering input: name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("Error registering input '%v': factory cannot be empty", name)
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf("Error registering input '%v': already registered", name)
	}

	registry[name] = factory
	logger.Info.Printf("Successfully registered input <%s>\n", name)

	return nil
}

func NewInput(content interface{}, docCh chan util.Doc) (Input, error) {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot convert given config to mapStr")
	}
	if len(contentMapStr) > 1 {
		return nil, fmt.Errorf("Cannot have multiple input targets.")
	}
	for target, val := range contentMapStr {
		if factory, ok := registry[target]; ok {
			input, err := factory(val, docCh)
			return input, err
		}
	}
	return nil, fmt.Errorf("Failed to creat any input target")
}

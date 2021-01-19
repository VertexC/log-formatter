package input

import (
	"fmt"

	"github.com/VertexC/log-formatter/connector"
	"github.com/VertexC/log-formatter/util"
)

type Input interface {
	Emit() map[string]interface{}
	Run()
}

type InputAgent struct {
	conn  *connector.Connector
	input Input
}

type Factory = func(interface{}) (Input, error)

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

func (agent *InputAgent) SetConnector(conn *connector.Connector) {
	agent.conn = conn
}

func (agent *InputAgent) SetConfig(content interface{}) error {
	// TODO: clean up resource of previous agent
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Cannot convert given config to mapStr")
	}
	if len(contentMapStr) > 1 {
		return fmt.Errorf("Cannot have multiple input targets.")
	}
	for target, val := range contentMapStr {
		var (
			input Input
			err   error
		)
		if factory, ok := registry[target]; ok {
			input, err = factory(val)
		} else if util.IsSoFile(target) {
			input, err = loadInputPlugin(target, val)
		} else {
			continue
		}
		if err != nil {
			return err
		} else {
			agent.input = input
			return nil
		}
	}
	return fmt.Errorf("Failed to creat any input target")
}

func loadInputPlugin(url string, content interface{}) (Input, error) {
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
		return nil, fmt.Errorf("`New` func in %s doesn't implement input interface", url)
	}
	instance, err := f(content)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (agent *InputAgent) Run() {
	agent.input.Run()
	go func() {
		for {
			agent.conn.InGate.Put(agent.input.Emit())
		}
	}()
}

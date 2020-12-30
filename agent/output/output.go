package output

import (
	"fmt"

	"github.com/VertexC/log-formatter/connector"
	"github.com/VertexC/log-formatter/util"
)

type Output interface {
	Run()
	// Send single doc to output
	Send(doc map[string]interface{})
}

type OutputAgent struct {
	conn   *connector.Connector
	output Output
}

type Factory = func(interface{}) (Output, error)

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

func (agent *OutputAgent) SetConnector(conn *connector.Connector) {
	agent.conn = conn
}

func (agent *OutputAgent) SetConfig(content interface{}) error {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Cannot convert given config to mapStr")
	}
	if len(contentMapStr) > 1 {
		return fmt.Errorf("Cannot have multiple output targets.")
	}
	for target, val := range contentMapStr {
		if factory, ok := registry[target]; ok {
			output, err := factory(val)
			if err == nil {
				agent.output = output
				return nil
			}
		}
	}
	return fmt.Errorf("Failed to creat any output target")
}

func (agent *OutputAgent) Run() {
	agent.output.Run()
	go func() {
		for {
			agent.output.Send(agent.conn.OutGate.Get())
		}
	}()
}

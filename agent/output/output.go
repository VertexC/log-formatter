package output

import (
	"fmt"

	"github.com/VertexC/log-formatter/agent/output/protocol"
	"github.com/VertexC/log-formatter/connector"
	"github.com/VertexC/log-formatter/util"
)

type OutputAgent struct {
	conn   *connector.Connector
	output protocol.Output
}

type Factory = func(interface{}) (protocol.Output, error)

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
		var (
			output protocol.Output
			err    error
		)
		if factory, ok := registry[target]; ok {
			output, err = factory(val)
		} else if util.IsSoFile(target) {
			output, err = loadOutputPlugin(target, val)
		} else {
			continue
		}
		if err != nil {
			return err
		} else {
			agent.output = output
			return nil
		}
	}
	return fmt.Errorf("Failed to creat any output target")
}

func loadOutputPlugin(url string, content interface{}) (protocol.Output, error) {
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
		return nil, fmt.Errorf("`New` func in %s doesn't implement output interface", url)
	}
	instance, err := f(content)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (agent *OutputAgent) Run() {
	agent.output.Run()
	go func() {
		for {
			agent.output.Send(agent.conn.OutGate.Get())
		}
	}()
}

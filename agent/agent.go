package agent

import (
	"fmt"

	"github.com/VertexC/log-formatter/agent/input"
	"github.com/VertexC/log-formatter/agent/output"
	"github.com/VertexC/log-formatter/agent/pipeline"
	"github.com/VertexC/log-formatter/config"
	"github.com/VertexC/log-formatter/connector"
	"github.com/VertexC/log-formatter/util"
)

const (
	Input    = "input"
	Output   = "output"
	Pipeline = "pipeline"
)

type Agent interface {
	Run()
	SetConnector(*connector.Connector)
	ChangeConfig(interface{}) error
}

type AgentsManager struct {
	BaseConfig config.ConfigBase
	agents     map[string]Agent
	logger     *util.Logger
}

func NewAgentsManager() (*AgentsManager, error) {
	logger := util.NewLogger("AgentsManager")

	conn, err := connector.NewConnector()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create connector: %s\n", err)
		logger.Error.Println(errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	manager := &AgentsManager{
		logger: logger,
		BaseConfig: config.ConfigBase{
			MandantoryFields: []string{Input, Output, Pipeline},
		},
		agents: map[string]Agent{
			Input:    new(input.InputAgent),
			Output:   new(output.OutputAgent),
			Pipeline: new(pipeline.PipelineAgent),
		},
	}

	for _, agent := range manager.agents {
		agent.SetConnector(conn)
	}

	return manager, nil
}

func (manager *AgentsManager) Run() {
	for _, agent := range manager.agents {
		agent.Run()
	}
}

func (manager *AgentsManager) ChangeConfigAndRun(content interface{}) error {
	// TODO: stop all agents

	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Cannot convert given config to mapStr")
	}

	manager.BaseConfig.Content = contentMapStr
	if err := manager.BaseConfig.Validate(); err != nil {
		err = fmt.Errorf("Config validation failed: %s\n", err)
		manager.logger.Error.Printf("%s\n", err)
		return err
	}

	for name, agent := range manager.agents {
		if err := agent.ChangeConfig(manager.BaseConfig.Content[name]); err != nil {
			err = fmt.Errorf("Failed to create %s: %s", name, err)
			manager.logger.Error.Printf("%s\n", err)
			return err
		}
	}

	for _, agent := range manager.agents {
		go agent.Run()
	}

	return nil
}

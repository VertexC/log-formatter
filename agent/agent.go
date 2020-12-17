package agent

import (
	"context"
	"fmt"
	"net"

	"github.com/VertexC/log-formatter/agent/input"
	"github.com/VertexC/log-formatter/agent/output"
	"github.com/VertexC/log-formatter/agent/pipeline"
	"github.com/VertexC/log-formatter/config"
	"github.com/VertexC/log-formatter/connector"
	agentpb "github.com/VertexC/log-formatter/proto/pkg/agent"
	"github.com/VertexC/log-formatter/util"

	"google.golang.org/grpc"
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
	agentpb.UnimplementedLogFormatterAgentServer

	Id     uint64
	Status agentpb.Status

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

func (manager *AgentsManager) ChangeConfig(context context.Context, request *agentpb.ChangeConfigRequest) (*agentpb.ChangeConfigResponse, error) {
	configBytes := request.Config
	manager.logger.Debug.Println(string(configBytes))
	return nil, nil
}

func (manager *AgentsManager) GetHeartBeat(context context.Context, request *agentpb.HeartBeatRequest) (*agentpb.HeartBeat, error) {
	manager.logger.Debug.Println("Got Heart Beat Get Request")
	msg := &agentpb.HeartBeat{
		Status: manager.Status,
		Id:     manager.Id,
	}
	return msg, nil
}

func (manager *AgentsManager) StartRpcService() {
	list, err := net.Listen("tcp", ":2001")

	if err != nil {
		manager.logger.Error.Fatalln("Failed to listen: %s", err)
	}
	s := grpc.NewServer()
	agentpb.RegisterLogFormatterAgentServer(s, manager)
	go func() {
		if err := s.Serve(list); err != nil {
			manager.logger.Error.Fatalln("Faied to server: %s", err)
		}
	}()
}

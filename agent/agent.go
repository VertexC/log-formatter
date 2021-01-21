package agent

import (
	"github.com/VertexC/log-formatter/agent/connector"
)

const (
	Input    = "input"
	Output   = "output"
	Pipeline = "pipeline"
)

type Agent interface {
	Run()
	Stop()
	SetConnector(*connector.Connector)
	SetConfig(interface{}) error
}

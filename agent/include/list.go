package include

import (
	// output plugins
	_ "github.com/VertexC/log-formatter/agent/output/console"
	_ "github.com/VertexC/log-formatter/agent/output/kafka"
	// input plugins
	_ "github.com/VertexC/log-formatter/agent/input/console"
	_ "github.com/VertexC/log-formatter/agent/input/elasticsearch"
	_ "github.com/VertexC/log-formatter/agent/input/kafka"
	// formatter plugins
	_ "github.com/VertexC/log-formatter/agent/pipeline/filter"
	_ "github.com/VertexC/log-formatter/agent/pipeline/forwarder"
	_ "github.com/VertexC/log-formatter/agent/pipeline/parser"
)

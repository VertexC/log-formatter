package include

import (
	// output plugins
	_ "github.com/VertexC/log-formatter/output/console"
	_ "github.com/VertexC/log-formatter/output/kafka"
	// input plugins
	_ "github.com/VertexC/log-formatter/input/console"
	_ "github.com/VertexC/log-formatter/input/kafka"
	// formatter plugins
	_ "github.com/VertexC/log-formatter/pipeline/filter"
	_ "github.com/VertexC/log-formatter/pipeline/forwarder"
	_ "github.com/VertexC/log-formatter/pipeline/parser"
)

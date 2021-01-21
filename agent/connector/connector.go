package connector

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	// export prometheus metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()
}

type Connector struct {
	InGate  *DocQueue
	OutGate *DocQueue
}

func NewConnector() (*Connector, error) {
	ig, err := NewDocQueue("in_gate")
	if err != nil {
		return nil, err
	}
	og, err := NewDocQueue("out_gate")
	if err != nil {
		return nil, err
	}

	connector := &Connector{
		InGate:  ig,
		OutGate: og,
	}
	return connector, nil
}

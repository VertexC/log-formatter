package connector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	qSize = 1000
)

var nameDict map[string]struct{}

//IOQueue wraps two i
type DocQueue struct {
	docCh       chan map[string]interface{}
	Name        string
	docProduced prometheus.Counter
	docConsumed prometheus.Counter
}

func NewDocQueue(name string) (*DocQueue, error) {
	if _, ok := nameDict[name]; ok {
		return nil, fmt.Errorf("DocQueue with name %s already exists", name)
	}
	q := &DocQueue{
		docCh: make(chan map[string]interface{}, qSize),
		Name:  name,
		docProduced: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("put_%s", name),
			Help: fmt.Sprintf("The total number of docs produced to DocQeue %s", name),
		}),
		docConsumed: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("get_%s", name),
			Help: fmt.Sprintf("The total number of docs consumed from DocQeue %s", name),
		}),
	}
	return q, nil
}

func (q *DocQueue) Put(doc map[string]interface{}) {
	defer q.docProduced.Inc()
	q.docCh <- doc
}

func (q *DocQueue) Get() map[string]interface{} {
	defer q.docConsumed.Inc()
	return <-q.docCh
}

func (q *DocQueue) GetCh() chan map[string]interface{} {
	return q.docCh
}

func (q *DocQueue) ConsumedInc() {
	q.docConsumed.Inc()
}

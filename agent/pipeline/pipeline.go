package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/VertexC/log-formatter/agent/pipeline/protocol"
	"github.com/VertexC/log-formatter/config"
	"github.com/VertexC/log-formatter/connector"
	"github.com/VertexC/log-formatter/util"
)

type worker struct {
	ctx    context.Context
	conn   *connector.Connector
	logger *util.Logger
	// TODO: move labelling to proper component of log-formatter
	labels     map[string]string
	formatters []protocol.Formatter
}

type PipelineConfig struct {
	Base   config.ConfigBase
	Worker int `yaml:"worker"`
}

type PipelineAgent struct {
	conn     *connector.Connector
	pipeline *Pipeline
	logger   *util.Logger
}

type Pipeline struct {
	logger  *util.Logger
	workers []*worker
	cancel  context.CancelFunc
	done    chan struct{}
}

var TAG = "PIPELINE"

func (agent *PipelineAgent) SetConnector(conn *connector.Connector) {
	agent.conn = conn
}

func (agent *PipelineAgent) SetConfig(content interface{}) error {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Failed to convert pipeline config to MapStr")
	}

	config := PipelineConfig{
		Base: config.ConfigBase{
			Content:          contentMapStr,
			MandantoryFields: []string{"formatters"},
		},
		Worker: 1,
	}

	if err := config.Base.Validate(); err != nil {
		return err
	}

	if err := util.YamlConvert(contentMapStr, &config); err != nil {
		return err
	}

	formatterCfgs, ok := contentMapStr["formatters"].([]interface{})
	if !ok {
		return fmt.Errorf("Failed to convert config to []MapStr")
	}

	ctx, cancel := context.WithCancel(context.Background())
	pipeline := new(Pipeline)
	pipeline.cancel = cancel
	pipeline.done = make(chan struct{})

	for i := 0; i < config.Worker; i++ {
		fmts := []protocol.Formatter{}
		for _, c := range formatterCfgs {
			fmt, err := NewFormatter(c)
			if err != nil {
				return err
			}
			fmts = append(fmts, fmt)
		}

		w := &worker{
			conn:       agent.conn,
			logger:     util.UseLog(TAG),
			formatters: fmts,
			ctx:        ctx,
		}
		pipeline.workers = append(pipeline.workers, w)
	}
	agent.pipeline = pipeline
	return nil
}

func (agent *PipelineAgent) Run() {
	go agent.pipeline.run()
}

func (agent *PipelineAgent) Stop() {
	util.UseLog(TAG).Info.Printf("Try to stop pipeline\n")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	wg := sync.WaitGroup{}
	wg.Add(1)
	defer cancel()
	go func() {
		defer wg.Done()
		for {
			if ctx.Err() != nil || len(agent.conn.InGate.GetCh()) == 0 {
				return
			}
		}
	}()
	wg.Wait()
	util.UseLog(TAG).Info.Printf("pipeline stopped\n")
}

func (agent *PipelineAgent) ChangeConfig(content interface{}) error {
	pipelineOld := agent.pipeline
	// if cannot create new pipeline from config, continue to run
	if err := agent.SetConfig(content); err != nil {
		return err
	}
	// new piepline has been created, stop the old pipeline
	pipelineOld.stop()
	util.UseLog(TAG).Info.Printf("Previous pipeline has stopped, start to run new pipeline\n")
	go agent.pipeline.run()
	return nil
}

func (pipeline *Pipeline) stop() {
	pipeline.cancel()
	<-pipeline.done
}

func (pipeline *Pipeline) run() {
	var wg sync.WaitGroup
	for _, worker := range pipeline.workers {
		wg.Add(1)
		go worker.run(&wg)
	}
	wg.Wait()
	pipeline.done <- struct{}{}
}

func (w *worker) run(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	ch := make(chan map[string]interface{})
	doFormat := func(doc map[string]interface{}) {
		discard := false
		for _, fmt := range w.formatters {
			var err error
			doc, err = fmt.Format(doc)
			if err != nil {
				discard = true
				w.logger.Warning.Printf("Discard doc:%s **with err** %s", doc, err)
			}
		}
		if !discard {
			for k, v := range w.labels {
				doc[k] = v
			}
			w.logger.Debug.Printf("%+v\n", doc)
			w.conn.OutGate.Put(doc)
		}
	}
	for {
		select {
		case doc := <-w.conn.InGate.GetCh():
			w.conn.InGate.ConsumedInc()
			doFormat(doc)
		case <-w.ctx.Done():
			w.logger.Info.Printf("Try to close a pipeline worker.\n")
			close(ch)
			return
		}
	}
}

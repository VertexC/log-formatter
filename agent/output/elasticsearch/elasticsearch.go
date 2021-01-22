package elasticsearch

import (
	"bytes"
	"math"
	"time"
	"context"
	"fmt"
	"encoding/json"

	"github.com/VertexC/log-formatter/util"
	"github.com/VertexC/log-formatter/logger"
	"github.com/VertexC/log-formatter/agent/output"
	"github.com/VertexC/log-formatter/agent/output/protocol"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"gopkg.in/yaml.v3"
)

type EsConfig struct {
	Host      string `yaml:"host"`
	Index     string `yaml:"index"`
	BatchSize int    `yaml:"batchsize"`
}

type EsOutput struct {
	logger      *logger.Logger
	docCh       chan map[string]interface{}
	config      EsConfig
	client      *elasticsearch.Client
	indexParser func(map[string]interface{}) string
}

func init() {
	output.Register("elasticsearch", New)
}

func New(content interface{}) (protocol.Output, error) {
	logger := logger.NewLogger("elastic-output")


	configMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get mapStr from config")
	}

	data, err := yaml.Marshal(&configMapStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to process given content as yaml: %s", err)
	}

	config := EsConfig{}
	yaml.Unmarshal(data, &config)

	// Declare an Elasticsearch configuration
	cfg := elasticsearch.Config{
		Addresses: []string{
			config.Host,
		},
	}

	// Instantiate a new Elasticsearch client object instance
	client, err := elasticsearch.NewClient(cfg)

	if err != nil {
		logger.Error.Fatalln("Elasticsearch connection error:", err)
	}

	// Have the client instance return a response
	if res, err := client.Info(); err != nil {
		logger.Error.Fatalln("client.Info() ERROR:", err)
	} else {
		logger.Info.Println("client response:", res)
	}
	output := &EsOutput{
		logger:      logger,
		docCh:       make(chan map[string]interface{}, 1000),
		config:      config,
		client:      client,
		indexParser: util.DynamicFromField(config.Index),
	}
	return output, nil
}

func (output *EsOutput) Send(doc map[string]interface{}) {
	output.docCh <- doc
}

func (output *EsOutput) Run() {
	logger := output.logger

	// Create a context object for the API calls
	ctx := context.Background()

	for {
		/*
			POST _bulk
			{ "create" : { "_index" : "test", "_id" : "3" } }
			{ "field1" : "value3" }
		*/

		batchSize := int(math.Max(100, float64(output.config.BatchSize)))
		logger.Trace.Println(batchSize)
		var bodyBuf bytes.Buffer
		startTime := time.Now()
		for {
			doc := <-output.docCh
			createLine := map[string]interface{}{
				"create": map[string]interface{}{
					"_index": output.indexParser(doc),
				},
			}
			if jsonStr, err := json.Marshal(createLine); err != nil {
				logger.Error.Fatalf("Failed to convert to json: %s\n", err)
			} else {
				bodyBuf.Write(jsonStr)
				bodyBuf.WriteByte('\n')
			}
			if jsonStr, err := json.Marshal(doc); err != nil {
				logger.Error.Fatalf("Failed to convert to json: %s\n", err)
			} else {
				bodyBuf.Write(jsonStr)
				bodyBuf.WriteByte('\n')
			}
			batchSize--
			// do request if batch is full or timeout
			if batchSize == 0 || time.Now().Sub(startTime).Seconds() >= float64(5) {
				break
			}
		}

		// batch update using bulk
		req := esapi.BulkRequest{
			Body:    &bodyBuf,
			Refresh: "true",
		}

		logger.Trace.Println(bodyBuf.String())

		go func() {
			// Return an API response object from request
			res, err := req.Do(ctx, output.client)
			if err != nil {
				logger.Error.Fatalf("IndexRequest ERROR: %s\n", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				logger.Error.Printf("ERROR indexing document with status: %s", res.Status())
			} else {
				var resMap map[string]interface{}
				decorder := json.NewDecoder(res.Body)
				if err := decorder.Decode(&resMap); err != nil {
					logger.Error.Printf("Error parsing the response body: %s\n", err)
				} else {
					// Print the response status and indexed document version.
					// logger.Trace.Printf("IndexRequest() RESPONSE: \nStatus: %s\n Result: %s\n Version:%s\n KvMap:%+v\n ",
					// 	res.Status(), resMap["result"], int(resMap["_version"].(float64)), resMap)
					logger.Trace.Printf("IndexRequest() RESPONSE: \nStatus: %s\n KvMap:%+v\n ",
						res.Status(), resMap)
				}
			}
		}()
	}
}

func (output *EsOutput) Stop() {}
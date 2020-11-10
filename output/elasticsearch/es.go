package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/VertexC/log-formatter/util"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type EsConfig struct {
	Host  string `yaml:"host"`
	Index string `yaml:"index"`
}

var logger = new(util.Logger)

func Execute(output EsConfig, outputCh chan interface{}, logFile string, verbose bool) {
	logger.Init(logFile, "Output-Es", verbose)

	// Create a context object for the API calls
	ctx := context.Background()

	// Declare an Elasticsearch configuration
	cfg := elasticsearch.Config{
		Addresses: []string{
			output.Host,
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

	for {
		/*
			POST _bulk
			{ "create" : { "_index" : "test", "_id" : "3" } }
			{ "field1" : "value3" }
		*/

		// FIXME: tailing messages will get blocked
		batchSize := 1000
		var bodyBuf bytes.Buffer
		for {
			kvMap := <-outputCh
			createLine := map[string]interface{}{
				"create": map[string]interface{}{
					"_index": output.Index,
				},
			}
			if jsonStr, err := json.Marshal(createLine); err != nil {
				logger.Error.Fatalf("Failed to convert to json: %s\n", err)
			} else {
				bodyBuf.Write(jsonStr)
				bodyBuf.WriteByte('\n')
			}

			if jsonStr, err := json.Marshal(kvMap); err != nil {
				logger.Error.Fatalf("Failed to convert to json: %s\n", err)
			} else {
				bodyBuf.Write(jsonStr)
				bodyBuf.WriteByte('\n')
			}
			batchSize--
			if batchSize == 0 {
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
			res, err := req.Do(ctx, client)
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

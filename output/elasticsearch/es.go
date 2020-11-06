package elasticsearch

import (
	"context"
	"encoding/json"
	"strings"

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
		kvMap := <-outputCh
		body, err := json.Marshal(kvMap)
		if err != nil {
			logger.Error.Printf("Failed to convert to json: %s\n", err)
			continue
		}
		// ask es to generate doc ID automatically
		docID := ""

		logger.Trace.Println(string(body))
		// Instantiate a request object
		req := esapi.IndexRequest{
			Index:      output.Index,
			DocumentID: docID,
			Body:       strings.NewReader(string(body)),
			Refresh:    "true",
		}

		// Return an API response object from request
		res, err := req.Do(ctx, client)
		if err != nil {
			logger.Error.Printf("IndexRequest ERROR: %s\n", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			logger.Error.Printf("%s ERROR indexing document ID=%d\n", res.Status(), docID)
		} else {
			// Deserialize the response into a map.
			var resMap map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&resMap); err != nil {
				logger.Error.Printf("Error parsing the response body: %s\n", err)
			} else {
				// Print the response status and indexed document version.
				logger.Trace.Printf("IndexRequest() RESPONSE: \nStatus: %s\n Result: %s\n Version:%s\n KvMap:%+v\n ",
					res.Status(), resMap["result"], int(resMap["_version"].(float64)), resMap)
			}
		}
	}
}

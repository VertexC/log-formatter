package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/VertexC/log-formatter/util"

	"github.com/elastic/go-elasticsearch/v8"
)

var logger = new(util.Logger)

type Query struct {
	Index string `yaml:"index"`
	Body  string `yaml:"body"`
	Retry int    `yaml:"retry" default:"0"` // retry qury in every <Retry>s
}

type EsConfig struct {
	Host   string  `yaml:"host"`
	Quries []Query `yaml:"quries"`
}

func Execute(input EsConfig, inputCh chan map[string]interface{}) {
	logger = util.NewLogger("Input-Es")

	var r map[string]interface{}

	// Initialize a client
	cfg := elasticsearch.Config{
		Addresses: []string{
			input.Host,
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logger.Error.Fatalf("Error creating the client: %s", err)
	}

	// Get cluster info
	res, err := es.Info()
	if err != nil {
		logger.Error.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		logger.Error.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		logger.Error.Fatalf("Error parsing the response body: %s", err)
	} else {
		body, _ := json.MarshalIndent(r, "", "  ")
		logger.Info.Println("client response:", string(body))
	}
	// Print client and server version numbers.
	logger.Info.Printf("Client: %s\n", elasticsearch.Version)
	logger.Info.Printf("Server: %s\n", r["version"].(map[string]interface{})["number"])

	// Build the request body.
	for _, query := range input.Quries {
		go func() {
			for {
				var buf bytes.Buffer

				if json.Valid([]byte(query.Body)) {
					buf.WriteString(query.Body)
				} else {
					logger.Error.Fatalf("Error encoding query %s to json\n", query.Body)
				}

				// Perform the search request.
				res, err = es.Search(
					es.Search.WithContext(context.Background()),
					es.Search.WithIndex(query.Index),
					es.Search.WithBody(&buf),
					es.Search.WithTrackTotalHits(true),
					es.Search.WithPretty(),
				)
				if err != nil {
					logger.Error.Fatalf("Error getting response: %s", err)
				}
				func() {
					defer res.Body.Close()

					if res.IsError() {
						var e map[string]interface{}
						if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
							logger.Error.Println(res.Body)
							logger.Error.Fatalf("Error parsing the response body: %s", err)
						} else {
							// Print the response status and error information.
							logger.Error.Fatalf("[%s] %s: %s",
								res.Status(),
								e["error"].(map[string]interface{})["type"],
								e["error"].(map[string]interface{})["reason"],
							)
						}
					}

					if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
						logger.Error.Fatalf("Error parsing the response body: %s", err)
					}
				}()

				// Print the response status, number of results, and request duration.
				logger.Trace.Printf(
					"[%s] %d hits; took: %dms",
					res.Status(),
					int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
					int(r["took"].(float64)),
				)
				// Print the ID and document source for each hit.
				for i, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
					logger.Trace.Printf("Return Id %d * ID=%s, %s", i, hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
					msg := hit.(map[string]interface{})["_source"]
					inputCh <- msg.(map[string]interface{})
				}

				logger.Trace.Println(strings.Repeat("=", 37))

				if query.Retry <= 0 {
					break
				}
				time.Sleep(time.Duration(query.Retry) * time.Second)
				logger.Info.Println("Retry after sleep")
			}
		}()
	}
}

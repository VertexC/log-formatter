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

type Query struct {
	Index string `yaml:"index"`
	Body  string `yaml:"body"`
	Retry int    `yaml:"retry" default:"0"` // retry qury in every <Retry>s
}

type EsConfig struct {
	Host   string  `yaml:"host"`
	Quries []Query `yaml:"quries"`
}

type EsInput struct {
	docCh  chan util.Doc
	config EsConfig
	logger *util.Logger
	es     *elasticsearch.Client
}

func NewEsInput(config EsConfig, docCh chan util.Doc) *EsInput {

	logger := util.NewLogger("elastic-input")

	var r util.Doc

	// Initialize a client
	cfg := elasticsearch.Config{
		Addresses: []string{
			config.Host,
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
	logger.Info.Printf("Server: %s\n", r["version"].(util.Doc)["number"])

	input := &EsInput{
		docCh:  docCh,
		config: config,
		logger: logger,
		es:     es,
	}

	return input
}

func (input *EsInput) Run() {

	logger := input.logger
	var r util.Doc
	// Build the request body.
	for _, query := range input.config.Quries {
		go func() {
			for {
				var buf bytes.Buffer

				if json.Valid([]byte(query.Body)) {
					buf.WriteString(query.Body)
				} else {
					input.logger.Error.Fatalf("Error encoding query %s to json\n", query.Body)
				}

				// Perform the search request.
				es := input.es
				res, err := es.Search(
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
						var e util.Doc
						if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
							logger.Error.Println(res.Body)
							logger.Error.Fatalf("Error parsing the response body: %s", err)
						} else {
							// Print the response status and error information.
							logger.Error.Fatalf("[%s] %s: %s",
								res.Status(),
								e["error"].(util.Doc)["type"],
								e["error"].(util.Doc)["reason"],
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
					int(r["hits"].(util.Doc)["total"].(util.Doc)["value"].(float64)),
					int(r["took"].(float64)),
				)
				// Print the ID and document source for each hit.
				for i, hit := range r["hits"].(util.Doc)["hits"].([]interface{}) {
					logger.Trace.Printf("Return Id %d * ID=%s, %s", i, hit.(util.Doc)["_id"], hit.(util.Doc)["_source"])
					msg := hit.(util.Doc)["_source"]
					input.docCh <- msg.(util.Doc)
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

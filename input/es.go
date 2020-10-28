package input

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type Query struct {
	Index     string `yaml:"index"`
	Body      string `yaml:"body"`
	Formatter string `yaml:"formatter"`
}

type InputConfig struct {
	Host   string  `yaml:"host"`
	Quries []Query `yaml:"quries"`
}

func EsSearch(input InputConfig, recordCh chan []interface{}, inLastJobCh chan int) {
	log.SetFlags(0)

	var (
		r map[string]interface{}
	)

	// Initialize a client with the default settings.
	//
	// An `ELASTICSEARCH_URL` environment variable will be used when exported.
	//
	cfg := elasticsearch.Config{
		Addresses: []string{
			input.Host,
		},
		// ...
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// 1. Get cluster info
	//
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	log.Printf("Client: %s", elasticsearch.Version)
	log.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
	log.Println(strings.Repeat("~", 37))

	// Build the request body.
	jobId := 0
	for _, query := range input.Quries {
		var buf bytes.Buffer

		if json.Valid([]byte(query.Body)) {
			buf.WriteString(query.Body)
		} else {
			log.Fatalf("Error encoding query: %s\n", err)
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
			log.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				log.Println(res.Body)
				log.Fatalf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				log.Fatalf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		}

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}
		// Print the response status, number of results, and request duration.
		log.Printf(
			"[%s] %d hits; took: %dms",
			res.Status(),
			int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(r["took"].(float64)),
		)
		// Print the ID and document source for each hit.
		for i, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			log.Printf("[Return Id %d] * ID=%s, %s", i, hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		}

		log.Println(strings.Repeat("=", 37))
		log.Printf("%+v\n", r)

		recordCh <- r["hits"].(map[string]interface{})["hits"].([]interface{})
		jobId++
	}
	inLastJobCh <- jobId
}

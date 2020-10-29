package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
	Default *log.Logger
)

type Query struct {
	Index     string `yaml:"index"`
	Body      string `yaml:"body"`
	Formatter string `yaml:"formatter"`
}

type EsConfig struct {
	Host   string  `yaml:"host"`
	Quries []Query `yaml:"quries"`
}

func Init() {
	file, err := os.OpenFile(path.Join("logs", "runtime.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	Trace = log.New(io.MultiWriter(file, os.Stdout),
		"[INPUT TRACE]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(io.MultiWriter(file, os.Stdout),
		"[INPUT INFO]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(io.MultiWriter(file, os.Stdout),
		"[INPUT WARNING]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(io.MultiWriter(file, os.Stderr),
		"[INPUT ERROR]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Debug = log.New(os.Stdout,
		"[INPUT DEBUG]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Default = log.New(io.MultiWriter(file, os.Stdout), "", 0)
}

func Execute(input EsConfig, recordCh chan []interface{}, inLastJobCh chan int) {
	Init()

	var r map[string]interface{}


	// Initialize a client
	cfg := elasticsearch.Config{
		Addresses: []string{
			input.Host,
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		Error.Fatalf("Error creating the client: %s", err)
	}

	// Get cluster info
	res, err := es.Info()
	if err != nil {
		Error.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		Error.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		Error.Fatalf("Error parsing the response body: %s", err)
	}
	// Print client and server version numbers.
	Info.Printf("Client: %s\n", elasticsearch.Version)
	Info.Printf("Server: %s\n", r["version"].(map[string]interface{})["number"])
	Default.Println(strings.Repeat("~", 37))

	// Build the request body.
	jobID := 0
	for _, query := range input.Quries {
		var buf bytes.Buffer

		if json.Valid([]byte(query.Body)) {
			buf.WriteString(query.Body)
		} else {
			Error.Fatalf("Error encoding query %s to json\n", query.Body)
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
			Error.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				Error.Println(res.Body)
				Error.Fatalf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				Error.Fatalf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		}

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			Error.Fatalf("Error parsing the response body: %s", err)
		}
		// Print the response status, number of results, and request duration.
		Trace.Printf(
			"[%s] %d hits; took: %dms",
			res.Status(),
			int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(r["took"].(float64)),
		)
		// Print the ID and document source for each hit.
		for i, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			Trace.Printf("Return Id %d * ID=%s, %s", i, hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		}

		Trace.Println(strings.Repeat("=", 37))

		recordCh <- r["hits"].(map[string]interface{})["hits"].([]interface{})
		jobID++
	}
	inLastJobCh <- jobID
}

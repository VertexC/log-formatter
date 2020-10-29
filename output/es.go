package output

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/VertexC/log-formatter/formatter"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
	Default *log.Logger
)

type OutputConfig struct {
	Host  string `yaml:"host"`
	Index string `yaml:"index"`
}

func Init() {
	file, err := os.OpenFile(path.Join("logs", "runtime.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	Trace = log.New(io.MultiWriter(file, os.Stdout),
		"[OUTPUT TRACE]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(io.MultiWriter(file, os.Stdout),
		"[OUTPUT INFO]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(io.MultiWriter(file, os.Stdout),
		"[OUTPUT WARNING]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(io.MultiWriter(file, os.Stderr),
		"[OUTPUT ERROR]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Debug = log.New(os.Stdout,
		"[OUTPUT DEBUG]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Default = log.New(io.MultiWriter(file, os.Stdout), "", 0)
}

func EsUpdate(output OutputConfig, recordCh chan []interface{}, outJobCh chan int) {
	// Create a context object for the API calls
	ctx := context.Background()

	// Declare an Elasticsearch configuration
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://es-qa.bjs.i.wish.com",
		},
	}

	// Instantiate a new Elasticsearch client object instance
	client, err := elasticsearch.NewClient(cfg)

	if err != nil {
		Error.Fatalln("Elasticsearch connection error:", err)
	}

	// Have the client instance return a response
	if res, err := client.Info(); err != nil {
		Error.Fatalln("client.Info() ERROR:", err)
	} else {
		Info.Println("client response:", res)
	}

	jobID := 0
	for {
		records := <-recordCh
		for _, record := range records {
			// Marshal Elasticsearch document struct objects to JSON string
			sourceMap := record.(map[string]interface{})["_source"].(map[string]interface{})
			message := sourceMap["message"].(string)
			_, labels, err := formatter.MongoFormatter(message)
			if err != nil {
				Error.Printf("Failed to format message %s, with error %s\n", message, err)
				continue
			}
			for key, val := range labels {
				sourceMap[key] = val
			}
			body, err := json.Marshal(sourceMap)
			if err != nil {
				Error.Printf("Failed to convert to json: %s\n", err)
				continue
			}
			// FIXME: change documentId as automatically genrated
			// or maybe use same Id as input
			docID := rand.Int()

			Info.Println(string(body))
			// Instantiate a request object
			req := esapi.IndexRequest{
				Index:      "bchen_playground",
				DocumentID: strconv.Itoa(docID),
				Body:       strings.NewReader(string(body)),
				Refresh:    "true",
			}

			// Return an API response object from request
			res, err := req.Do(ctx, client)
			if err != nil {
				Error.Printf("IndexRequest ERROR: %s\n", err)
			}
			defer res.Body.Close()

			if res.IsError() {
				Error.Printf("%s ERROR indexing document ID=%d\n", res.Status(), docID)
			} else {
				// Deserialize the response into a map.
				var resMap map[string]interface{}
				if err := json.NewDecoder(res.Body).Decode(&resMap); err != nil {
					Error.Printf("Error parsing the response body: %s\n", err)
				} else {
					Trace.Printf("IndexRequest() RESPONSE:")
					// Print the response status and indexed document version.
					Trace.Println("Status:", res.Status())
					Trace.Println("Result:", resMap["result"])
					Trace.Println("Version:", int(resMap["_version"].(float64)))
					Trace.Println("resMap:", resMap)
				}
			}
		}
		jobID++
		outJobCh <- jobID
	}
}

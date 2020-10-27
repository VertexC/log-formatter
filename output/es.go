package output

import (
	"log"
	"fmt"
	"strings"
	"strconv"
	"context"
	"encoding/json"
	"math/rand"

	"github.com/VertexC/log-formatter/formatter"
	
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func EsUpdate(records []interface{}) {

    // Allow for custom formatting of log output
    log.SetFlags(0)

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
        fmt.Println("Elasticsearch connection error:", err)
    }

    // Have the client instance return a response
    if res, err := client.Info(); err != nil {
        log.Fatalf("client.Info() ERROR:", err)
    } else {
        log.Printf("client response:", res)
	}
	
	for _, record := range records {
		fmt.Println(record)
		// Marshal Elasticsearch document struct objects to JSON string
		sourceMap := record.(map[string]interface{})["_source"].(map[string]interface{})
		message := sourceMap["message"].(string)
		_, labels := formatter.MongoFormatter(message)
		for key, val := range labels {
			sourceMap[key] = val
		}
		body, err := json.Marshal(sourceMap)
		if err != nil {
			log.Fatal("Failed to convert to json:", err)
		}
		// FIXME: change documentId as automatically genrated
		docId := rand.Int()
		
		fmt.Println(string(body))
		// Instantiate a request object
		req := esapi.IndexRequest{
			Index:      "bchen_playground",
			DocumentID: strconv.Itoa(docId),
			Body:       strings.NewReader(string(body)),
			Refresh:    "true",
		}

		// Return an API response object from request
		res, err := req.Do(ctx, client)
		if err != nil {
			log.Fatalf("IndexRequest ERROR: %s", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			log.Printf("%s ERROR indexing document ID=%d", res.Status(), docId)
		} else {
			// Deserialize the response into a map.
			var resMap map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&resMap); err != nil {
				log.Printf("Error parsing the response body: %s", err)
			} else {
				log.Printf("\nIndexRequest() RESPONSE:")
				// Print the response status and indexed document version.
				fmt.Println("Status:", res.Status())
				fmt.Println("Result:", resMap["result"])
				fmt.Println("Version:", int(resMap["_version"].(float64)))
				fmt.Println("resMap:", resMap)
				fmt.Println("\n")
			}
		}
	}
}
package formatter

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
)

var Version = "0.0.0"

// MongoLog is a template which consists components of mongdb log messages
// https://docs.mongodb.com/v3.2/replication/
type MongoLog struct {
	timestamp string
	severity  string
	component string
	context   string
	message   string
}

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}
	return subMatchMap
}

// GetLabelsMango extract interested label from mongodb message based on specific context
// TODO: convert this into config
func GetLabelsMango(message string, component string) map[string]interface{} {
	labels := map[string]interface{}{}

	fmt.Println(component)
	switch component {
	case "COMMAND":
		// dbname and command type at the begining of message
		match := regexp.MustCompile(`command\s+(?P<dbname>.*?)\s+command:\s+(?P<command>.*?)\s+{`).FindStringSubmatch(message)
		labels["dbname"] = match[1]
		labels["command"] = match[2]
		// protocal and time at the end of message
		match = regexp.MustCompile(`protocol\:(?P<protocal>.*?)\s+(?P<time>\d+)ms`).FindStringSubmatch(message)
		labels["protocal"] = match[1]
		if time, err := strconv.ParseFloat(match[2], 32); err != nil {
			log.Fatal("Cannot parse time field to float32!")
		} else {
			labels["time"] = time
		}
		// TODO: try to parse the inner json-like body to json
		// planSummary(optional)
		match = regexp.MustCompile(`planSummary:\s+(?P<plan>.*?)\s+{`).FindStringSubmatch(message)
		if len(match) != 0 {
			labels["plan"] = match[1]
		}
	}
	return labels
}

// MongoFormatter designed to parse mongodb log message (from 3.2 to 4.3)
// TODO: validate the ealieast version that fits
// it returns a mongo
// <timestamp> <severity> <component> [<context>] <message>
func MongoFormatter(msg string) (*MongoLog, map[string]interface{}) {
	regex := `(?P<timestamp>\d{4}-\d{2}-\d{2}T\d{2}\:\d{2}\:\d{2}.\d+(?:\+|-)\d+)` // for timestamp in iso8601-local, which is default
	regex += `\s+` + `(?P<serverity>(?:F|E|W|I|D))`
	regex += `\s+` + `(?P<component>(?:ACCESS|COMMAND|CONTROL|ELECTION|FTDC|GEO|INDEX|INITSYNC|NETWORK|QUERY|REPL|REPL_HB|ROLLBACK|SHARDING|STORAGE|RECOVERY|JOURNAL|TXN|WRITE)?)` // TODO: add other component type
	regex += `\s+` + `(?P<context>\[.*?\])`
	regex += `\s+` + `(?P<message>.*$)`
	// fmt.Println(regex)
	re := regexp.MustCompile(regex)
	matchMap := reSubMatchMap(re, msg)

	// Use relect to fill up fields
	mongoLog := MongoLog{}

	mongoLog.timestamp = matchMap["timestamp"]
	mongoLog.severity = matchMap["serverity"]
	mongoLog.component = matchMap["component"]
	mongoLog.context = matchMap["context"]
	mongoLog.message = matchMap["message"]

	labels := GetLabelsMango(mongoLog.message, mongoLog.component)

	fmt.Printf("Labels: %+v\n", labels)
	fmt.Printf("MangoLog: %+v\n", mongoLog)
	return &mongoLog, labels
}

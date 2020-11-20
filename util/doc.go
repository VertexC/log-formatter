package util

import (
	"reflect"
)

// Doc is a wrap type over map[string]interface{} with other functionalities
type Doc map[string]interface{}

var (
	endDoc Doc = nil
)

func IsEndDoc(doc Doc) bool {
	return reflect.DeepEqual(doc, endDoc)
}

func GetEndDoc() Doc {
	return endDoc
}

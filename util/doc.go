package util

import (
	"reflect"
)

var (
	endDoc map[string]interface{} = nil
)

func IsEndDoc(doc map[string]interface{}) bool {
	return reflect.DeepEqual(doc, endDoc)
}

func EndDoc() map[string]interface{} {
	return endDoc
}

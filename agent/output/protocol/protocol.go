package protocol

type Output interface {
	Run()
	// Send single doc to output
	Send(doc map[string]interface{})
}

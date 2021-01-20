package protocol

type Output interface {
	Run()
	Stop()
	// Send single doc to output
	Send(doc map[string]interface{})
}

package protocol

type Input interface {
	Emit() map[string]interface{}
	Run()
}

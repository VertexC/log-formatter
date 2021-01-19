package protocol

type Formatter interface {
	Format(map[string]interface{}) (map[string]interface{}, error)
}

type Factory = func(interface{}) (Formatter, error)

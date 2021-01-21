The main purpose of log-formatter is to process log data from Kafka and forward results to various outputs. 

Though it supports other inputs including elasticsearch and file, it is stateless, which means it won't store any information like Filebeat. Also, it assumes one-to-one relation between input and output, the format of input data is assumed to be consistent (a topic in Kafka), which makes log-formatter easier to scale.

The general implementation of log-formatter consists of three parts: `input`, `pipeline`, `output`. Log data from input is converted to `Doc`, alias of `map[string]interface{}`. `Doc` goes through a pipeline of multiple formatters, which may `add,delete,update` the fields in `Doc`. Finally `Doc` is then converted to the specific schema and forward to outputs.

## Releases
(TODO:)

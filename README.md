# Log Formatter
Log Formatter provides configurable pipeline to process log data.

``` Raw Logs (string) -> KV Maps -> Enhanced KV Maps```

# Usage
```bash
-c string
    config file path (default "config.yml")
-cpuprof
    enable cpu profile
-memprof
    enable mem profile
-v    add TRACE/WARNING logging if enabled
```

# Configuration
**A Example Config**: Request mongo log from an es server, parse it with general-formatter with labels, send enhanced results to another es server. 
```yaml
log: "logs"
output:
  target: "elasticsearch"
  elasticsearch:
    host: <host>
    index: mongodb-log-formatted
input:
  target: "elasticsearch"
  elasticsearch:
    host: <host>
    quries:
      - index: mongodb-log*
        retry: 30
        body: '{
          "query": {
            "range": {
              "@timestamp": {
                "gt": "now-30s/s", 
                "lt": "now/s"
              }
            }
          }
        }'
formatter: 
  type: "general"
  general:
    components: (?P<timestamp>\d{4}-\d{2}-\d{2}T\d{2}\:\d{2}\:\d{2}.\d+(?:\+|-)\d+)\s+(?P<serverity>(?:F|E|W|I|D))\s+(?P<component>(?:[A-Z]+)?)\s+\[(?P<context>.*?)\]\s+(?P<message_>.*$) 
    labels:
      - component: message_
        regexprs:
          - command\s+(?P<namespace>.*?)\s+command:\s+(?P<comand>.*?)\s+
          - protocol:(?P<protocal>.*?)\s+(?P<time>\d+)ms
          - \$comment\:\s+\"(?P<ip_>\[.*?\])(?P<pyFunc>.*?)\s+\@\s+(?P<pyFile>.*?\.py:[0-9]+)\"
          - planSummary:\s+(?P<planSummary>[A-Z]+\s+\{.*?\})\s+

```
###  Formatter Plugins
- General Formatter
  - Parse log into different components
  - Enhance logs with labels (Add fiels to KV Map).

### Input Plugins
- Kafka
- Elasticsearch
### Output Plugins
- Console
- Elasticsearch


package filter

import (
	"fmt"
	"regexp"

	"github.com/VertexC/log-formatter/agent/pipeline"
	"github.com/VertexC/log-formatter/util"
)

type FilterConfig struct {
	IncludeFields []string `yaml:"include_fields"`
	ExcludeFields []string `yaml:"exclude_fields"`
	// TODO: to accelerate filtering fixed fields
	// with cache set, generate white/black list for filtering rest docs
	// Cache bool `yaml:"cache"`
}

type Filter struct {
	IncludeRegex []*regexp.Regexp
	ExcludeRegex []*regexp.Regexp
	logger       *util.Logger
}

func init() {
	pipeline.Register("filter", New)
}

func New(content interface{}) (pipeline.Formatter, error) {
	contentMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to convert config to MapStr")
	}
	config := FilterConfig{}
	if err := util.YamlConvert(contentMapStr, &config); err != nil {
		return nil, err
	}

	filter := new(Filter)
	filter.logger = util.NewLogger("filter")
	if len(config.IncludeFields) > 0 && len(config.ExcludeFields) > 0 {
		return nil, fmt.Errorf("Cannot use include and exlude at same time")
	}
	for _, regStr := range config.IncludeFields {
		r := regexp.MustCompile(regStr)
		filter.IncludeRegex = append(filter.IncludeRegex, r)
	}

	for _, regStr := range config.ExcludeFields {
		r := regexp.MustCompile(regStr)
		filter.ExcludeRegex = append(filter.ExcludeRegex, r)
	}
	return filter, nil
}

func (filter *Filter) Format(doc map[string]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for field, val := range doc {
		for _, r := range filter.IncludeRegex {
			if r.MatchString(field) {
				result[field] = val
				break
			}
		}
	}
	return result, nil
}

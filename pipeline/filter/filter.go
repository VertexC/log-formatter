package filter

import (
	"errors"
	"github.com/VertexC/log-formatter/util"
	"regexp"
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

func NewFilter(config FilterConfig) (*Filter, error) {
	filter := new(Filter)
	filter.logger = util.NewLogger("filter")
	if len(config.IncludeFields) > 0 && len(config.ExcludeFields) > 0 {
		return nil, errors.New("Cannot use include and exlude at same time")
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

func (filter *Filter) Format(doc util.Doc) (util.Doc, error) {
	result := util.Doc{}
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

package filter_test

import (
	"github.com/VertexC/log-formatter/pipeline/filter"
	"github.com/VertexC/log-formatter/util"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

// default config for test cases
var defaultConfigYml = `
include_fields:
  - u.*
  - messagXXX
`
var doc = map[string]interface{}{"message": "123 [HELLO] hello foo", "user": "vertexc"}

func init() {
	util.LogFile = ""
}

func createDefaultConfig() filter.FilterConfig {
	return filter.FilterConfig{
		IncludeFields: []string{"u.*", "messagXXX"},
	}
}

func TestNewFromConfig(t *testing.T) {
	config := new(filter.FilterConfig)
	err := yaml.Unmarshal([]byte(defaultConfigYml), config)
	if err != nil {
		assert.Fail(t, "Cannot parse from yaml")
	}
	defaultCfg := createDefaultConfig()
	assert.Equal(t, defaultCfg, *config)
}

func TestParseWithDefaultConfig(t *testing.T) {
	expectedDoc := map[string]interface{}{
		"user": "vertexc",
	}
	defaultCfg := createDefaultConfig()
	fmt, err := filter.NewFilter(defaultCfg)
	assert.Equal(t, err, nil)
	result, err := fmt.Format(doc)
	assert.Equal(t, err, nil)
	assert.Equal(t, expectedDoc, result)
}

var invalidConfigYml = `
include_fields:
- u.*
- messagXXX
exclude_fields:
- xxx
`

func TestNewFromInvalidConfig(t *testing.T) {

	config := new(filter.FilterConfig)
	err := yaml.Unmarshal([]byte(invalidConfigYml), config)
	if err != nil {
		assert.Fail(t, "Cannot parse from yaml")
	}
	fmt, err := filter.NewFilter(*config)
	assert.NotNil(t, err)
	assert.Nil(t, fmt)
}

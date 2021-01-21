package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigBase struct {
	Content          map[string]interface{}
	MandantoryFields []string
}

// Validate validates the mandantory fields within content of ConfigBase
func (c *ConfigBase) Validate() error {
	for _, field := range c.MandantoryFields {
		if _, ok := c.Content[field]; !ok {
			return fmt.Errorf("Failed to Validate Config: <%s> not exists", field)
		}
	}
	return nil
}

// GetMapStr return the val
func (c *ConfigBase) GetMapStr(field string) (map[string]interface{}, error) {
	if _, ok := c.Content[field]; ok {
		if val, ok := c.Content[field].(map[string]interface{}); ok {
			return val, nil
		} else {
			return nil, fmt.Errorf("Failed to convert <%s>'s val to mapStr", field)
		}
	} else {
		return nil, fmt.Errorf("Field <%s> not exists", field)
	}
}

//  LoadMapStrFromYamlFile load content from yaml file and unmarhsal it
func LoadMapStrFromYamlFile(url string) (map[string]interface{}, error) {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("Frailed to get config from %s: %s", url, err)
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("Frailed to get config from %s: %s", url, err)
		}
		config, err := LoadMapStrFromYamlBytes(data)
		return config, err
	}
	// other wise, load from local file
	data, err := ioutil.ReadFile(url)
	if err != nil {
		return nil, err
	}
	config, err := LoadMapStrFromYamlBytes(data)
	return config, err
}

// LoadMapStrFromYamlBytes load config from bytes and unmarshal it as yaml to MapStr
func LoadMapStrFromYamlBytes(data []byte) (map[string]interface{}, error) {
	config := map[string]interface{}{}
	err := yaml.Unmarshal(data, &config)
	return config, err
}

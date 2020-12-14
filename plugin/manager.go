package plugin

import (
	"fmt"
	"plugin"
	"regexp"

	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/util"
)

// IsSoFile check if a given name has .so in the end
func IsSoFile(name string) bool {
	r := regexp.MustCompile("^*.so$")
	return r.MatchString(name)
}

// LoadFormatterPlugin try to loads formatter from .so file
// return name of plugin if successful
func LoadFormatterPlugin(pluginPath string, content interface{}, docCh chan map[string]interface{}) (input.Input, error) {
	// try to loadPlugin from file
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("Could not open %s: %s", pluginPath, err)
	}
	newFunc, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("could not find New function in %s: %s", pluginPath, err)
	}
	f, ok := newFunc.(input.Factory)
	if !ok {
		return nil, fmt.Errorf("`New` func in %s doesn't implement interface", pluginPath)
	}
	instance, err := f(content, docCh)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

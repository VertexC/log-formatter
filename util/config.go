package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v3"
)

// DynamicFromField return a `function` with based on given str
// if str consist xxx{{field}}xxx, the `function` will replace
// `field` with doc[field]
// otherwise, the `function` simply return s
func DynamicFromField(s string) func(Doc) string {
	// {{field}}
	regexpStr := `\{\{(?P<index>.*?)\}\}`
	r := regexp.MustCompile(regexpStr)
	matchMap, err := SubMatchMapRegex(regexpStr, s)
	if err == nil {
		if token, exist := matchMap["index"]; exist {
			return func(doc Doc) string {
				index := doc[token].(string)
				return r.ReplaceAllString(s, index)
			}
		}
	}
	return func(doc Doc) string {
		return s
	}
}

type Fragment struct {
	content *yaml.Node
}

func (f *Fragment) UnmarshalYAML(value *yaml.Node) error {
	var err error
	// process includes in fragments
	f.content, err = resolveIncludes(value)
	return err
}

type IncludeProcessor struct {
	Target interface{}
}

func (i *IncludeProcessor) UnmarshalYAML(value *yaml.Node) error {
	resolved, err := resolveIncludes(value)
	if err != nil {
		return err
	}
	return resolved.Decode(i.Target)
}

func resolveIncludes(node *yaml.Node) (*yaml.Node, error) {
	if node.Tag == "!include" {
		if node.Kind != yaml.ScalarNode {
			return nil, errors.New("!include on a non-scalar node")
		}
		file, err := ioutil.ReadFile(node.Value)
		if err != nil {
			return nil, err
		}
		var f Fragment
		err = yaml.Unmarshal(file, &f)
		return f.content, err
	}
	if node.Kind == yaml.SequenceNode || node.Kind == yaml.MappingNode {
		var err error
		for i := range node.Content {
			node.Content[i], err = resolveIncludes(node.Content[i])
			if err != nil {
				return nil, err
			}
		}
	}
	return node, nil
}

// YamlConvert try to convert contentMapStr to target with yaml (un)marshal
func YamlConvert(contentMapStr map[string]interface{}, target interface{}) error {
	data, err := yaml.Marshal(contentMapStr)
	if err != nil {
		return fmt.Errorf("Failed to convert with yaml: %s", err)
	}

	err = yaml.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("Failed to convert to yaml: %s", err)
	}

	return nil
}

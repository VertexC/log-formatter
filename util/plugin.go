package util

import (
	"io"
	"net/http"
	"os"
	"plugin"
	"regexp"
	"strings"
)

// IsSoFile check if a given name has .so in the end
func IsSoFile(name string) bool {
	r := regexp.MustCompile("^*.so$")
	return r.MatchString(name)
}

func LoadPlugin(url string) (*plugin.Plugin, error) {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		filePath := "tmp.so"
		out, err := os.Create(filePath)
		_, err = io.Copy(out, resp.Body)
		p, err := plugin.Open(filePath)
		return p, err
	}
	// otherwise, try to load it as local file
	p, err := plugin.Open(url)
	return p, err
}

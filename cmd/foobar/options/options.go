package options

import (
	"io/ioutil"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Options struct {
}

func Load(f string) (c *Options, err error) {
	c = &Options{}
	err = parseConfigFile(f, c)
	return
}

func parseConfigFile(f string, out interface{}) (err error) {
	filename, err := filepath.Abs(f)
	if err != nil {
		return
	}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	return yaml.Unmarshal(yamlFile, out)
}

package main

import (
	"io/ioutil"
	"launchpad.net/goyaml"
)

// A Config represents the key-value pairs in a _config.yml file.
// The file is freeform, and thus requires the flexibility of a map.
type Config map[string]interface{}

// Sets a parameter value.
func (c Config) Set(key string, val interface{}) {
	c[key] = val
}

// Gets a parameter value.
func (c Config) Get(key string) interface{} {
	return c[key]
}

// ParseConfig will parse a YAML file at the given path and return
// a key-value Config structure.
//
// ParseConfig always returns a non-nil map containing all the
// valid YAML parameters found; err describes the first unmarshalling
// error encountered, if any.
func ParseConfig(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseConfig(b)
}

func parseConfig(data []byte) (Config, error) {
	conf := map[string] interface{} { }
	err := goyaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

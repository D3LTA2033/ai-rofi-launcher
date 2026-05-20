package main

import "gopkg.in/yaml.v3"

func marshalYAML(c Config) (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

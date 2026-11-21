package main

import (
	"log"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Matches []ConfigMatch
}

type ConfigMatch struct {
	Name  string
	Route ConfigRoute
	Out   []map[string]any
}

type ConfigRoute struct {
	Path   string
	Method string
}

func Discover(file string) Config {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}
	var cfg Config
	if err := yaml.Load(data, &cfg); err != nil {
		log.Fatalf("error parsing yaml: %v", err)
	}
	return cfg
}

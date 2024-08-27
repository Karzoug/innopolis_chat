package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var cfgFilename = "config.yaml"

type Config struct {
	Token   string
	Address string
}

func mustLoadConfig() Config {
	f, err := os.Open(cfgFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	return cfg
}

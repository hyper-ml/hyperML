package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	UserID            string
	DefaultServerAddr string
	SessionId         string
	Concurrency       int
}

var configDir = filepath.Join(os.Getenv("HOME"), ".hyperflow")
var configPath = filepath.Join(configDir, "config.json")

func ReadFromFile() (*Config, error) {
	var config *Config

	config_encoded, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Default(), nil
	}
	err = json.Unmarshal(config_encoded, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func Default() *Config {
	return &Config{
		DefaultServerAddr: "http://localhost:8888",
		Concurrency:       10,
	}
}

func GetConfig() *Config {

	c, err := ReadFromFile()

	if err != nil {
		base.Error("Failed to read config file")
		c = Default()
	}
	return c
}

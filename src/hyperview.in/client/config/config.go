package config

import (
  "os" 
  "fmt"
  "io/ioutil"
  "encoding/json"
  "path/filepath"
)

type Config struct {
  UserID string
  ServerAddr string
  SessionId string
  RepoUriPath string
  ObjectUriPath string
  Concurrency int
}

var configDir = filepath.Join(os.Getenv("HOME"), ".hyperview")
var configPath = filepath.Join(configDir, "config.json")

func ReadFromFile() (*Config, error) {
  var config *Config

  config_encoded, err := ioutil.ReadFile(configPath)
  if err != nil {
    fmt.Println("No Config file at %s", configPath)
    return Default(), nil
  }
  err = json.Unmarshal(config_encoded, &config)
  if err != nil {
    return nil, err
  }

  return config, nil
}

func Default() (*Config) {
  return &Config{
    ServerAddr: "http://localhost:8888",
    RepoUriPath: "repo",
    ObjectUriPath: "object",
    Concurrency: 10,
  }
}




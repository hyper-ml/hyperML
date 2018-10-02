package config

import (
  "os"
  "fmt"
  "io/ioutil"
  "encoding/json"
  "path/filepath"

  "hyperview.in/server/base"
)

const (
  ConfigPathEnvName ="HFLOW_CONFIG_PATH"
  ConfigFilePerm = 0644
)


type Config struct {
  Interface string
  DbConfig *DbConfig
  StorageOption string
  S3 *S3Config
  Gcs *GcsConfig

  LogLevel string
  LogPath string
}


type DbConfig struct {
  Driver string // POSTGRES, ETCD
  DbName string
  Dbuser string
  Dbpass string
}

type S3Config struct {
  CredPath string
  AccessKey string
  SecretKey string
  Bucket string
}


type GcsConfig struct {
  CredPath string
  Bucket string
}

func SetConfigPath(v string) error { 
  return base.SetEnvVar(ConfigPathEnvName, v)
}


func GetConfigPath() string  { 
  path := base.GetEnvVar(ConfigPathEnvName)
  
  if path == "" {
    path = filepath.Join(base.HomeDir(), ".hflow")
    base.SetEnvVar(ConfigPathEnvName, path)
    base.Log("Chooising default Config path: ", path)
  }

  return path
}

func GetConfig() (*Config, error) {
  return readFromFile()
}


func readFromFile() (*Config, error) {
  var config *Config

  config_path := GetConfigPath()

  if config_path == "" {
    return nil, fmt.Errorf("Config path not set")
  }

  config_json, err := ioutil.ReadFile(config_path)
  if err != nil {
    fmt.Println("Failed to read config file at: ", config_path)
    return nil, err
  }

  err = json.Unmarshal(config_json, &config)
  if err != nil {
    return nil, err
  }

  return config, nil
}

func UpdateConfig(c *Config) error {
  return writeToFile(c)
}

func SetConfig(c *Config) (error) {
  return writeToFile(c)
}

func writeToFile(config *Config) (error) {
  config_path := GetConfigPath()
  if config_path == "" {
    return nil, fmt.Errorf("Config path not set")
  }

  json_config, err := json.Marshal(config)
  err = ioutil.WriteFile(config_path, json_config, ConfigFilePerm)

  return err
}




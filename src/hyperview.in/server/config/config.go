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
  GStorage    = "GCS" 
  AwsStorage       = "S3"
)


const (
  ConfigPathEnvName ="HFLOW_CONFIG_PATH"
  ConfigFilePerm = 0644
)


type Config struct {
  Interface string
  DB *DbConfig
  StorageOption string
  S3 *S3Config
  Gcs *GcsConfig

  LogLevel string
  LogPath string
}


type DbConfig struct {
  Driver string // POSTGRES, ETCD
  Name string
  User string
  Pass string
  File string
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

func defaultConfig() *Config {
  return &Config{
    DbConfig: &DbConfig{},
    S3: &S3Config{},
    GCS: &GCSConfig{},
  }
}

func SetConfigPath(v string) error { 
  return base.SetEnvVar(ConfigPathEnvName, v)
}


func GetConfigPath() string  { 
  path := base.GetEnvVar(ConfigPathEnvName)
  
  if path == "" {
    home_dir, _ := base.HomeDir()
    path = filepath.Join(home_dir, ".hflow")
    if err := base.SetEnvVar(ConfigPathEnvName, path); err != nil {
      base.Log("Failed to set config path", err)
    }

    base.Log("Choosing default Config path: ", path) 
  }

  return path
}

func GetConfig() (*Config, error) {
  config_path := GetConfigPath()

  if config_path == "" {
    return nil, fmt.Errorf("Config path not found")
  }

  config, err := readFromFile(config_path)
  
  if err != nil {
    if os.IsNotExist(err) { 
      fmt.Println("Config file doesnt exist. So creating a new one at: ", config_path)
      return createConfig(config_path)
    }
    return nil, err
  }

  return config, err
}

func createConfig(path string) (*Config, error) {
  default_config := defaultConfig()
  json_config, _ := json.Marshal(default_config)
  err := ioutil.WriteFile(path, json_config, ConfigFilePerm)
  if err != nil {
    base.Error("Failed to create default config file: ", err)
    return nil, err
  }
  return default_config, nil
}

func readFromFile(path string) (*Config, error) {
  var config *Config

  config_json, err := ioutil.ReadFile(path)
  if err != nil {
    fmt.Println("Failed to read config file at: ", path)
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
  if config_path == ""{
    return fmt.Errorf("Config path not set")
  }

  json_config, err := json.Marshal(config)
  err = ioutil.WriteFile(config_path, json_config, ConfigFilePerm)

  return err
}




package config

import (
  "os" 
  "fmt"
  "io/ioutil"
  "encoding/json"
  "path/filepath"

  "hyperflow.in/server/pkg/base"
)
 
const (
  ConfigFilePerm = 0644 
  DefaultLocalStorage = "/var/tmp/hflow"
) 

type Config struct {
  DefaultServerAddr string
  Jwt string
  UserName string
  Email string
  Concurrency int 
  Local *Local
}

type Local struct {
  ObjStorage string  
  StorageLimit int32
}

var home = base.JustHomeDir() 
var configPath = filepath.Join(home, ".hflow_user")

func ReadFromFile() (*Config, error) {
  var config *Config

  config_encoded, err := ioutil.ReadFile(configPath)
  if err != nil {
    if os.IsNotExist(err) { 
      default_config :=Default()
      return default_config, nil
    }
    return nil, err
  }
  
  err = json.Unmarshal(config_encoded, &config)
  if err != nil {
    return nil, err
  }

  return config, nil
}

func Default() *Config {

  // create obj storage   

  config := &Config {
    DefaultServerAddr: "http://localhost:8888",
    Concurrency: 10, 
    Local: {
      ObjStorage: DefaultLocalStorage,
    },
  }
  return config
}

func GetConfig() *Config {

  c, err := ReadFromFile()

  if err != nil {
    fmt.Println("Failed to read config file")
    c = Default();
  } 

  return c
}

func createConfigFile() (*Config, error) {
  return configToFile(Default())
}

func configToFile(cfg *Config) (c *Config, fnError error) {

  json_obj, _ := json.Marshal(cfg)
  err := ioutil.WriteFile(configPath, json_obj, ConfigFilePerm)

  if err != nil {
    base.Error("Failed to create config file: ", err)
    return nil, err
  }

  return cfg, nil
}

func writeConfig(param, value string ) (*Config, error) {
  var cfg *Config
  var err error
  json_obj, err := ioutil.ReadFile(configPath)
  if err != nil {
    if !os.IsNotExist(err) {
      return nil, err
    }
  }

  if json_obj != nil {
    if err := json.Unmarshal(json_obj, &cfg); err != nil {
      return nil, err
    }
  } else {
    cfg, err = createConfigFile() 
    if err != nil {
      return nil, err
    }
  }

  switch param {
  case "DefaultServerAddr":
    cfg.DefaultServerAddr = value
  case "Jwt":
    cfg.Jwt = value
  case "Email":
    cfg.Email = value
  case "UserName":
    cfg.UserName = value 
  case "ObjStorage":
    cfg.Local.ObjStorage = value
  }

  if cfg, err = configToFile(cfg); err != nil {
    return nil, err
  }

  return cfg, err
}

func SetHostServer(addr string) error {
  _, err := writeConfig("DefaultServerAddr", addr)
  return err
} 

func SetEmail(email string) error {
  _, err := writeConfig("Email", email)
  return err
}

func SetJwt(jwt string) error {
  _, err := writeConfig("Jwt", jwt)
  return err
}

package config

import (
  "os" 
  "fmt"
  "io/ioutil"
  "encoding/json"
  "path/filepath"
)

type UrlMap struct {
  RepoUriPath string
  RepoAttrsUriPath string
  DatasetUriPath string
  BranchAttrsUriPath string
  CommitUriPath string
  CommitAttrsUriPath string
  CommitMapUriPath string
  FileUriPath string
  FileAttrsUriPath string
  ObjectUriPath string
  FlowUriPath string
  FlowAttrsUriPath string
  TaskAttrsUriPath string
  TaskStatusUriPath string
  WorkerUriPath string
  // virtual file system to uri
  VfsUriPath string
}

type Config struct {
  UserID string
  DefaultServerAddr string
  SessionId string 
  Concurrency int
  UrlMap *UrlMap
}

var configDir = filepath.Join(os.Getenv("HOME"), ".hyperview")
var configPath = filepath.Join(configDir, "config.json")

func ReadFromFile() (*Config, error) {
  var config *Config

  config_encoded, err := ioutil.ReadFile(configPath)
  if err != nil {
    fmt.Println("No Config file at :", configPath)
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
    DefaultServerAddr: "http://localhost:8888",
    Concurrency: 10,
    UrlMap: &UrlMap {
      RepoUriPath: "repo",
      RepoAttrsUriPath: "repo_attrs",
      DatasetUriPath: "dataset",
      BranchAttrsUriPath: "branch_attr",
      CommitUriPath: "commit",
      CommitAttrsUriPath: "commit_attrs",
      CommitMapUriPath: "commit_map",
      FileUriPath: "file",
      FileAttrsUriPath: "file_attrs",
      ObjectUriPath: "object",
      VfsUriPath: "vfs",
      WorkerUriPath: "worker",
      FlowUriPath: "flow",
      FlowAttrsUriPath: "flow_attrs",
      TaskAttrsUriPath: "tasks",
      TaskStatusUriPath: "task_status",
      }, 
  }
}

func GetConfig() *Config {

  c, err := ReadFromFile()

  if err != nil {
    fmt.Println("Failed to read config file")
    c = Default()
  } 
  return c
}


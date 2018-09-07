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
  DefaultServerAddr string
  SessionId string
  RepoUriPath string
  RepoAttrsUriPath string
  DataSetUriPath string
  BranchAttrsUriPath string
  CommitUriPath string
  CommitAttrsUriPath string
  CommitMapUriPath string
  FileUriPath string
  FileAttrsUriPath string
  ObjectUriPath string
  FlowAttrsUriPath string
  TaskAttrsUriPath string
  TaskStatusUriPath string
  WorkerUriPath string
  // virtual file system to uri
  VfsUriPath string
  Concurrency int
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
    RepoUriPath: "repo",
    RepoAttrsUriPath: "repo_attrs",
    DataSetUriPath: "dataset",
    BranchAttrsUriPath: "branch_attr",
    CommitUriPath: "commit",
    CommitAttrsUriPath: "commit_attrs",
    CommitMapUriPath: "commit_map",
    FileUriPath: "file",
    FileAttrsUriPath: "file_attrs",
    ObjectUriPath: "object",
    VfsUriPath: "vfs",
    WorkerUriPath: "worker",
    FlowAttrsUriPath: "flow",
    TaskAttrsUriPath: "tasks",
    TaskStatusUriPath: "task_status",
    Concurrency: 10,
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


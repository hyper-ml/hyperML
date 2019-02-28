package utils

import (
  "os"
  "fmt"
  "strings"

  "path/filepath"

  "hyperflow.in/client"
  "hyperflow.in/client/config"
)
const (
  DefaultRepoDirPerm = 0755
  ModelSubPath = "saved_models"
  OutSubPath = "out"
  DataSubPath = "dataset"
  IgnoreFileName = ".hfignore"
  ListSeparator = "\n"
)


func ExitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}

func check_exists(dir string) bool {
  return DirExists(dir)
}

func createComponentRepo(repo_type, dir, parent_name string) error {
  r, err := config.ReadRepoParams(dir, "REPO_NAME")
  if r != "" {
    return nil
  }  

  c, _ := client.New(dir) 
  switch repo_type {
    case "MODEL":
      if err := c.InitModelRepo(dir, parent_name); err != nil {
        return err
      }
    case "DATASET":
      if err := c.InitDataRepo(dir, parent_name); err != nil {
        return err
      }
    case "OUT":
      if err := c.InitOutRepo(dir, parent_name); err != nil {
        return err
      }
    default: 
      if err := c.InitRepo(parent_name); err != nil {
        return err
      }
  }
  
  err = config.WriteRepoParams(dir, "REPO_NAME", parent_name)
  err = config.WriteRepoParams(dir, "BRANCH_NAME", "master")

  return err
}

func createModelRepo(dir, parent_name string) error {
  return createComponentRepo("MODEL", dir, parent_name + "/model")
}

func createDataRepo(dir, parent_name string) error {
  return createComponentRepo("DATASET", dir, parent_name + "/dataset")
}

func createOutRepo(dir, parent_name string) error {
  return createComponentRepo("OUT", dir, parent_name + "/out")
}

func createIgnoreFile(repo_path string) error {
  file_path := filepath.Join(repo_path, IgnoreFileName)
  if !PathExists(file_path) {
    default_ignores := []byte("saved_models\nout\ndataset\n")
    return CreateFileContent(file_path, default_ignores, 0)
  }

  return nil
}

func CreateComponentDirs(repo_path, repo_name string) error {

  model_dir := filepath.Join(repo_path, ModelSubPath)

  if !check_exists(model_dir) {
    if err := MkDirAll(model_dir, DefaultRepoDirPerm); err != nil {
      return err 
    }
  }
  
  if err := createModelRepo(model_dir, repo_name); err != nil {
    return err
  }

  out_dir := filepath.Join(repo_path, OutSubPath)

  if !check_exists(out_dir) {
    if err := MkDirAll(out_dir, DefaultRepoDirPerm); err != nil {
      return err 
    }
  }
  
  if err := createOutRepo(out_dir, repo_name); err != nil {
    return err
  }

  data_dir := filepath.Join(repo_path, DataSubPath)

  if !check_exists(data_dir) {
    if err := MkDirAll(data_dir, DefaultRepoDirPerm); err != nil {
      return err 
    }
  }
  
  if err := createDataRepo(data_dir, repo_name); err != nil {
    return err
  }

  if err := createIgnoreFile(repo_path); err != nil {
    return err
  }

  return nil
}
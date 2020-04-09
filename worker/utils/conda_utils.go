package utils

import (
  "fmt"
  "os/exec"
  "io/ioutil"
  "encoding/json"
  "path/filepath"
  "gopkg.in/yaml.v2"
)
 
type envList struct {
  Envs []string
}

type envSpec struct {
  Name string
}

func checkConda() error {
  cmd := exec.Command("conda", "--help")
  err := cmd.Run()
  if err != nil {
    return fmt.Errorf("failed to locate conda installation: " + err.Error())
  }
  return nil
}

func createCondaEnvWithSpec(spec string) error {
  cmd := exec.Command("conda", "env", "create", "-q", "-f", spec)
  out, err := cmd.CombinedOutput()

  if err != nil {
    return fmt.Errorf("failed to create conda environment: \n" + string(out) + "\n \n" + err.Error())
  }

  return nil
}

func createCondaPythonEnv(name string) error {
  cmd := exec.Command("conda", "create","-y", "-q", "-n", name, "python")
  out, err := cmd.CombinedOutput()

  if err != nil {
    return fmt.Errorf("failed to create conda environment: \n" + string(out) + "\n \n" + err.Error())
  }
  
  return nil
}



func checkCondaEnv(name string) (bool, error) {

  if name == "" {
    return false, fmt.Errorf("No environment found in yaml spec.")
  }

  env_list, err := exec.Command("conda", "env", "list","--json").Output()
  if err != nil {
    return false, err
  }

  var elist envList
  err = json.Unmarshal(env_list, &elist) 
  if err != nil {
    return false, err
  }

  for _, e := range elist.Envs { 
    if filepath.Base(e) == name {
      return true, nil
    }
  }

  return false, nil
}


func readYaml(path string) (*envSpec, error) {
  spec_data, err := ioutil.ReadFile(path)
  if err != nil {
    return nil, err
  }

  env_spec := envSpec{}
  err = yaml.Unmarshal([]byte(spec_data), &env_spec)
  if err != nil {
    return nil, err
  }
  return &env_spec, nil
}

func GetOrCreateDefaultCondaEnv(name string) (string, error) {
  var env_name string
  var err error

  if err := checkConda(); err !=nil {
    return env_name, err
  }

  exst, err := checkCondaEnv(name) 
  if err != nil {
    return env_name, err
  }

  if !exst {
    if err := createCondaPythonEnv(name); err != nil {
      return env_name,err
    }
  }

  return env_name, nil
}

func GetOrCreateCondaEnvWithSpec(spec string) (string, error) {
  var envName string

  if err := checkConda(); err !=nil {
    return envName, err
  }

  env_spec, err := readYaml(spec)
  if err != nil {
    return envName, err
  }
  
  envName = env_spec.Name
  exst, err := checkCondaEnv(envName) 
  if err != nil {
    return envName, err
  }

  if !exst {
    if err := createCondaEnvWithSpec(spec); err != nil {
      return envName, err
    }
  }

  return envName, nil
}








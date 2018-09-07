package config

import (
  "os"
  "io/ioutil"
  "path/filepath"
  "fmt"
  "encoding/json"
  "github.com/spf13/viper"

  "hyperview.in/server/base"

)
const (
  DefaultRepoParamsFileName = ".hflow_repo"
)

func errMandatoryParam(s string) error {
  return fmt.Errorf("Missing a mandatory parameter: %s", s)
}
 
func GetRepoConfig(dir string) (*viper.Viper, error) {
  return GetOrCreateRepoConfig(dir, false)
}

func GetOrCreateRepoConfig(dir string, createIfMissing bool) (*viper.Viper, error) {
  // read 
  if dir == "" {
    return nil, errMandatoryParam("dir")
  }

  file_name := DefaultRepoParamsFileName 
  param_path := dir + "/" + file_name

  param_reader, err := os.Open(param_path)
  if err != nil {

    if (!createIfMissing) {
      return nil, fmt.Errorf("Failed to read repo parameters file")
    }

    base.Log("failed to find repo params. error: ", err)
    err := defaultRepoParams(dir)
    if err != nil {
      base.Log("Failed to create default repo params: ", err)
      return nil, fmt.Errorf("Failed to create repo params")
    }
  }

  param_reader, err = os.Open(param_path)

  v := viper.New()
  v.SetConfigType("json")
  
  err = v.ReadConfig(param_reader) // Find and read the config file
  if err != nil {
    return nil, err
  }

  return v, nil 
}

// writes default param json file to given directory with filename
func defaultRepoParams(dir string) error {
  // creates default param file in workspace directory
  // values: repo_id, commit_id, task_id, current_branch
  v := viper.New() 
  file_name := DefaultRepoParamsFileName
  v.SetConfigName(file_name)
  v.SetConfigType("json")

  v.Set("REPO_NAME", "")
  v.Set("COMMIT_ID", "")
  v.Set("TASK_ID", "")
  v.Set("CURRENT_BRANCH", "")
  v.Set("API_SERVER", "")

  c := v.AllSettings()
  b, _ := json.MarshalIndent(c, "", "  ")

  p:= filepath.Join(dir, file_name) 
  err:= ioutil.WriteFile(p,  b, 0644)

  if err != nil {
    base.Log("error writing default config file: ", err)
    return err
  }

  return nil
}



func WriteRepoParams(working_dir string, key string, value string) error {
  v, _ := GetOrCreateRepoConfig(working_dir, true)
  v.Set(key, value)

  c := v.AllSettings()
  b, _ := json.MarshalIndent(c, "", "  ")

  p:= filepath.Join(working_dir, DefaultRepoParamsFileName) 
  err:= ioutil.WriteFile(p,  b, 0644)

  if err != nil {
    base.Log("error writing config file: ", err)
    return err
  }

  return nil
}


func ReadRepoParams(working_dir string, key string) (value string, err error) {
  v, _ := GetRepoConfig(working_dir)
  //base.Log("repo_id: " , v.Get(key))
  vi := v.Get(key)
  return vi.(string), nil
}
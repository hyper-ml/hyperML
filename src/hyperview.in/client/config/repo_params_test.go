package config_test

import(
  "fmt"
  "testing"
  "hyperview.in/client/config"
)

const (
  TEST_WORKING_DIR = "/Users/apple/MyProjects/stash/one_repo"
)

func Test_SetRepoParams(t *testing.T) {
  v, _ := config.GetRepoConfig(TEST_WORKING_DIR)
  fmt.Println("v: ", v.AllSettings())
}


func Test_ReadRepoParams(t *testing.T) {
  id, _:= config.ReadRepoParams(TEST_WORKING_DIR, "repo_id")
  fmt.Println("id: ", id)

  c_id, _:= config.ReadRepoParams(TEST_WORKING_DIR, "commit_id")
  fmt.Println("c_id: ", c_id)
}


func Test_WriteRepoParams(t *testing.T) {
  _ = config.WriteRepoParams(TEST_WORKING_DIR, "repo_id", "23dser")
  
  id, _:= config.ReadRepoParams(TEST_WORKING_DIR, "repo_id")
  fmt.Println("id: ", id)

  c_id, _:= config.ReadRepoParams(TEST_WORKING_DIR, "commit_id")
  fmt.Println("c_id: ", c_id)
}
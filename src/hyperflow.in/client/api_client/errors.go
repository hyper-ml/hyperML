package api_client

import (
  "fmt"
)

func pullRepoError(err error) error {
  return fmt.Errorf("pull_repo_error: ", err.Error())
}

func unknownError(s string) error {
  return fmt.Errorf("Unknown error: " + s)
}
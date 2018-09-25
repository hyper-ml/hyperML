package api_client

import (
  "fmt"
)

func pullRepoError(err error) error {
  return fmt.Errorf("pull_repo_error: ", err.Error())
}
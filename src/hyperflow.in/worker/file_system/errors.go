package file_system

import (
  "fmt"
)

func errCommitStatus(status string) error {
  return fmt.Errorf("invalid_commit_status: %s", status)
}

func errInvalidRepo() error {
  return fmt.Errorf("invalid_repo: Repo is invalid")
}
package workspace


import (
  "fmt"
)

func errInvalidRepoName(name string) error{
  return fmt.Errorf("invalid_repo_name: %s", name)
}

func errInvalidCommitId() error {
  return fmt.Errorf("invalid_commit_id: Invalid Commit Id")
}

func errStaleCommit() error {
  return fmt.Errorf("stale_commit: You have older version of repo. Please pull the repo again.")
}

func errBranchMissing(name string) error {
  return fmt.Errorf("branch_missing: %s", name)
}
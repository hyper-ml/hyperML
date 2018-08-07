package file_system

import (
  "testing"
  "hyperview.in/worker"
)



func Test_repoFsMount(t *testing.T) {
  wc, _ := worker.NewWorkerClient("")

  repoFs:= NewRepoFs("/Users/apple/MyProjects/stash/hw", 0, "test_repo", "7759ff96ae2046bbbf89de9bf14ab381", wc)
  _ = repoFs.Mount()

}


package file_system

import (
  "fmt"
  "io/ioutil"
  
  "testing"
  "hyperview.in/worker"

)

const (
  TEST_REPO_NAME ="test_repo"
  TEST_WS_LOCAL_PATH = "/Users/apple/MyProjects/stash/hw"
  TEST_WS_OUT_PATH ="/out"
  TEST_COMMIT_ID = "7759ff96ae2046bbbf89de9bf14ab381"
)

func onError(t *testing.T, fnName string, err error){
  if err != nil {
    fmt.Println("[" +fnName + "] Error: ", err)
    t.Fail()
  }
}

func Test_RepoFsMount(t *testing.T) {
  wc, _ := worker.NewWorkerClient("")

  repoFs:= NewRepoFs(TEST_WS_LOCAL_PATH, 0, TEST_REPO_NAME, TEST_COMMIT_ID, wc)
  _ = repoFs.Mount()

}

//TODO: create an object and then push
func Test_RepoFsCommitOutputs(t *testing.T) {
  wc, _ := worker.NewWorkerClient("")
  repoFs:= NewRepoFs(TEST_WS_LOCAL_PATH, 0, TEST_REPO_NAME, TEST_COMMIT_ID, wc)

  d1 := []byte("hello\ngo\n")
  err := ioutil.WriteFile(TEST_WS_LOCAL_PATH + TEST_WS_OUT_PATH + "/t3.txt", d1, 0644)
  onError(t, "Test_RepoFsPushObject", err)

  err = ioutil.WriteFile(TEST_WS_LOCAL_PATH + TEST_WS_OUT_PATH + "/t4.txt", d1, 0644)
  onError(t, "Test_RepoFsPushObject", err)

  upld_size, err:= repoFs.commitOutPath()
  onError(t, "Test_RepoFsPushObject", err)

  
  fmt.Println("Upload size: ", upld_size)
}

 
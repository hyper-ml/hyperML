package code_sync

import(
  "fmt"
  "testing"  
  "hyperview.in/client/fs"

  ws "hyperview.in/server/core/workspace"
)


const (
  TEST_REPO_NAME = "test_repo"
  TEST_COMMIT_ID = "b2874bdc9cfd4d0cacd8b8091e90fa93"
  TEST_FILE_NAME = "file"
  TEST_REPO_DIR = "/Users/apple/MyProjects/stash"
  TEST_SERVER_ADDR = "http://localhost:8888"
)

func Test_GetFileByUrl(t *testing.T) { 
  // load local FS 
  localFs := fs.NewFS(TEST_REPO_DIR, 0)

  c := NewClient(TEST_SERVER_ADDR, localFs)
  s, err:= c.PullObject(TEST_REPO_NAME, TEST_COMMIT_ID, TEST_FILE_NAME)
  fmt.Println("s:", s)
  if err != nil {
    t.Fatalf("Failed to get URL: %s", err)
  }
}

func Test_PullRemoteRepo(t *testing.T) { 
  // load local FS 
  localFs := fs.NewFS(TEST_REPO_DIR, 0)

  c := NewClient(TEST_SERVER_ADDR, localFs)
  f_map := make(map[string]ws.File)
  f_map["file"] = ws.File{}

  preq := &PullRepoRequest{
    CommitId: TEST_COMMIT_ID,
    RepoName: TEST_REPO_NAME,
    FileMap: f_map,
  }

  s, err:= c.PullRemoteRepo(preq, 1)
  fmt.Println("s:", s)
  if err != nil {
    t.Fatalf("Failed to get URL: %s", err)
  }
}
package worker



import (
  "testing"
  "fmt"  
  "encoding/json" 
  "bytes"
  "hyperview.in/worker/schema"

)

const (
  TEST_REPO_DIR = "/Users/apple/MyProjects/stash"
  TEST_REPO_NAME = "test_repo"
)

func Test_NewWorkerClient(t *testing.T) {
  _, err:= NewWorkerClient()
  if err != nil {
    fmt.Println("failed to create ApiClient")
    t.Fatalf("Failed to create ApiClient")
  }
}



func Test_GetRepo(t *testing.T) {
  client, err:= NewWorkerClient()
  if err != nil {
    fmt.Println("failed to create ApiClient", err)
    t.Fatalf("Failed to create ApiClient")
  }

  repo_request:= client.RepoInfo.Verb("GET")
  repo_request.Param("repoName", TEST_REPO_NAME)
  result := repo_request.Do() 

  getRepoResponse := schema.GetRepoResponse{}
  body, err := result.Raw()
  if (err != nil) {
    fmt.Println("failed to get raw result:", result)
    t.Fatalf("failed to get raw result")
  }

  err = json.Unmarshal(body, &getRepoResponse)
  if err != nil {
    fmt.Println("failed to unmarshal result:", getRepoResponse)
    t.Fatalf("failed to unmarshal getRepoResponse")
  }

  if (getRepoResponse.Repo.Name != TEST_REPO_NAME) {
    t.Fatalf("Failed to fetch Repo")
  }
}
 
func Test_Lookup(t *testing.T) {
  client, err := NewWorkerClient()
  if err != nil {
    fmt.Println("failed to create ApiClient", err)
    t.Fatalf("Failed to create ApiClient")
  }

  _, err = client.LookupFile(TEST_REPO_NAME, "", "/workspace")
  if err != nil {
    fmt.Println("failed to lookup", err)
    t.Fatalf("Failed to lookup")    
  }
}


func Test_GetFileObject(t *testing.T) {
  client, err := NewWorkerClient()
  if err != nil {
    fmt.Println("failed to create ApiClient", err)
    t.Fatalf("Failed to create ApiClient")
  }
  
  var buff bytes.Buffer
  err =client.GetFileObject(TEST_REPO_NAME, "b2874bdc9cfd4d0cacd8b8091e90fa93","file", 0, 5, &buff)
  if err!= nil {
    t.Fatalf("Failed to get object: %s", err)
  }
  fmt.Println("output of get file ob:", buff.String())
}


func Test_PutFileObject(t *testing.T) {
   client, err := NewWorkerClient()
  if err != nil {
    fmt.Println("failed to create ApiClient", err)
    t.Fatalf("Failed to create ApiClient")
  }

  err = client.PutFileObject()

}

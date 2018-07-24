package client



import (
  "testing"
  "fmt" 
  "encoding/json" 
  "hyperview.in/client/schema"

)

const (
  TEST_REPO_DIR = "/Users/apple/MyProjects/stash"
  TEST_REPO_NAME = "test_repo"
)

func Test_NewApiClient(t *testing.T) {
  _, err:= NewApiClient(TEST_REPO_DIR)
  if err != nil {
    fmt.Println("failed to create ApiClient")
    t.Fatalf("Failed to create ApiClient")
  }

}

func Test_GetRepo(t *testing.T) {
  client, err:= NewApiClient(TEST_REPO_DIR)
  if err != nil {
    fmt.Println("failed to create ApiClient", err)
    t.Fatalf("Failed to create ApiClient")
  }

  repo_request:= client.repoInfo.Verb("GET")
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

//TODO: create a repo and then clone and then delete 
func Test_CloneRepo(t *testing.T) {
  client, err := NewApiClient(TEST_REPO_DIR)
  if err != nil {
    fmt.Println("failed to create ApiClient", err)
    t.Fatalf("Failed to create ApiClient")
  }

  err = client.CloneRepo(TEST_REPO_NAME)
  if err != nil {
    fmt.Println("failed to clone Repo", err)
    t.Fatalf("Failed to clone repo")    
  }
}

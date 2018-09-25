package api_client



import (
  "testing"
  "fmt" 
  "encoding/json" 
  "hyperview.in/client/schema"

)

const (
  TEST_REPO_DIR = "/var/tmp/one_repo"
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


func Test_RequestLog(t *testing.T) {
  client, err := NewApiClient(TEST_REPO_DIR)
  if err != nil {
    fmt.Println("failed to create ApiClient", err)
    t.Fatalf("Failed to create ApiClient")
  }

  _, err = client.RequestLog("3336dc6d88d24e3c936d925bb3fc0cf1")
  if err != nil {
    fmt.Println("failed to clone Repo", err)
    t.Fatalf("Failed to clone repo")    
  }
}  

func Test_InitDataRepo(t *testing.T) {
  c, err := NewApiClient(TEST_REPO_DIR)
  err = c.InitDataRepo(TEST_REPO_DIR, "TestdataRepo001")
  if err != nil {
    fmt.Println("Failed to create data repo: ", err)
    t.Fail()
  }
} 
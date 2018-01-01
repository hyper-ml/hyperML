package rest_client


import (
	"testing"
  "fmt"
  "encoding/json"
  "net/url"
  
)

const (
  TEST_REPO_NAME = "test_repo"
  TEST_HOST_STRING = "http://localhost:8888"
  TEST_REPO_PATH = "repo"
)

func Test_NewRequest(t *testing.T) {
  url_string:= TEST_HOST_STRING + "/"+ TEST_REPO_PATH + "?repoName=" + TEST_REPO_NAME
  fmt.Println("Accessing url:", url_string)

  url_object, err := url.Parse(url_string) 

  if err != nil {
    fmt.Println("failed to generate URL", err)
  }
  request := NewRequest(nil, "GET", url_object, "", 0 )
  request.Param("repoName", TEST_REPO_NAME)
  result := request.Do()


  getRepoResponse := getRepoResponse{}
  // capture errors from body 
  fmt.Println("result.body:", string(result.body))  

  err = json.Unmarshal(result.body, &getRepoResponse)

  fmt.Println("result:", getRepoResponse.Repo.Name)

}

package workspace


import (
  "fmt"
	"testing"
  //"encoding/json"
  //"strings" 
  "hyperview.in/server/core/db"

  "hyperview.in/server/core/storage"
)

const (
  TEST_DB_USER = "apple"
  TEST_DB_PASSWORD = ""
  TEST_DB_NAME = "amp_db"
  TEST_REPO_NAME = "test_repo"
  TEST_DIR = "test_dir"
) 


func getDb() (*db.DatabaseContext, error) {
  conn, err := db.NewDatabaseContext(TEST_DB_NAME, TEST_DB_USER, TEST_DB_PASSWORD)

  return conn, err
}
  
func getObjApi() (storage.ObjectAPIServer, error) {
  return storage.NewObjectAPI(TEST_DIR, 0, "GCS") 
}

func getAPIServer() (*apiServer, error) {
  dbc, err := getDb()

  if err != nil { 
    return nil, err
  }

  objApi, err := getObjApi()
  if err != nil {
    fmt.Println("objApi Error:", err)
    return nil, err
  }
  fmt.Println("objApi:",objApi)

  api, err := NewApiServer(dbc, objApi)
  fmt.Println("api:", api)
  if err != nil {
    return nil, err
  }

  return api, err
}

func Test_StartRepo(t *testing.T) {

  dbc, err := getDb()

  if err != nil {
    fmt.Println("DB Error:", err)
    t.Fatalf("Authentication Error in initialization of DB")
  }

  api, err := NewApiServer(dbc, nil)
  if err != nil {
    fmt.Println("Unable to create API Server ")
    t.Fatalf("Unable to create API Server")
  }
  err = api.RemoveRepo(TEST_REPO_NAME)

  err = api.InitRepo(TEST_REPO_NAME)

  if err != nil {
    fmt.Println("Start Repo Error", err)
    t.Fatalf("Unable to start or create repo")
  }
}

func Test_GetRepo(t *testing.T) {

  dbc, err := getDb()

  if err != nil {
    fmt.Println("DB Error:", err)
    t.Fatalf("Authentication Error in initialization of DB")
  }

  objApi, err := getObjApi()
  if err != nil {
    fmt.Println("objApi Error:", err)
    t.Fatalf("Error in initialization of ObjectAPIServer")
  }

  api, err := NewApiServer(dbc, objApi)
  if err != nil {
    fmt.Println("Unable to create API Server ")
    t.Fatalf("Unable to create API Server")
  }

  _, err = api.GetRepo(TEST_REPO_NAME)

  if err != nil {
    fmt.Println("Get Repo Error", err)
    t.Fatalf("Unable to get repo")
  }
}


/*
func Test_Commit(t *testing.T) {
  api, _ := getAPIServer()
  
  _, err := api.InitCommit(TEST_REPO_NAME,"master")
  if err != nil {
    fmt.Println("Init commit error :", err)
    t.Fatalf("Failed to initialize commit")
  }

  err = api.CloseCommit(TEST_REPO_NAME)
  if err != nil {
    fmt.Println("Finish commit error :", err)
    t.Fatalf("Failed to complete commit")
  }  
}


func Test_AddFile(t *testing.T) {

  api, err := getAPIServer()
  if err != nil {
    fmt.Println("Unable to create API Server ")
    t.Fatalf("Unable to create API Server")
  }


  _, err = api.AddFileToRepo(TEST_REPO_NAME, "file", strings.NewReader("Hello Worlddd"))
  if err != nil {

    fmt.Println("Add file to repo Error", err)
    t.Fatalf("Unable to add file repo")
  }

}
func Test_InitCommit(t *testing.T) {
  api, _ := getAPIServer()
  _, err := api.InitCommit(TEST_REPO_NAME, "master")
  if err != nil {
    fmt.Println("InitCommit commit error :", err)
    t.Fatalf("InitCommit to complete commit")
  }  
}

func Test_CloseCommit(t *testing.T) {
  api, _ := getAPIServer()
  err := api.CloseCommit(TEST_REPO_NAME)
  if err != nil {
    fmt.Println("Finish commit error :", err)
    t.Fatalf("Failed to complete commit")
  }  
}


func Test_DownloadRepo(t *testing.T) {
  api, _ := getAPIServer()
  commit_attrs, err := api.DownloadRepo(TEST_REPO_NAME)  
  if err != nil {
    fmt.Println("Failed to retrieve repo", err)
    t.Fatalf("Failed to retrieve repo")
  } 
  b, _ := json.Marshal(commit_attrs)
  fmt.Println("commit_attrs:", string(b)) 

}  */

package workspace


import (
  "fmt"
	"testing"
  //"encoding/json"
  //"strings" 
  "hyperview.in/server/core/db"
  "hyperview.in/server/core/utils"
  "hyperview.in/server/base"

  "hyperview.in/server/core/storage"
)

const (
  TEST_DB_USER = "apple"
  TEST_DB_PASSWORD = ""
  TEST_DB_NAME = "amp_db"
  TEST_REPO_NAME = "test_repo"
  TEST_BRANCH_NAME = "master"
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

  _, err = api.InitRepo(TEST_REPO_NAME)

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


func Test_GetOutRepo(t *testing.T) {
  // delete test repo
  // create test repo 

  // get out repo 
  // delete out repo 
}

func Test_GetModelRepo(t *testing.T) {
  // delete test repo
  db, err := utils.FakeDb()
  test_name := TEST_REPO_NAME
  test_branch := TEST_BRANCH_NAME

  api_server, err := NewApiServer(db, nil)
  if err != nil {
    base.Log("[Test_GetModelRepo] Failed to create API Server: ", err)
    t.Fatalf("Failed to create API Server")
  }

  _ = api_server.RemoveRepo(test_name)

  // create test repo 
  _, err = api_server.InitRepo(test_name)
  if err != nil {
    base.Log("[Test_GetModelRepo] Failed to create test repo: ", err)
    t.Fatalf("Failed to create a test repo")
  }

  // create a commit 
  commit_attrs, _ := api_server.InitCommit(test_name, test_branch, "")

  // get model repo 
  model_repo, err := api_server.GetOrCreateModelRepo(test_name, test_branch, commit_attrs.Commit.Id)
  if err != nil {
    t.Fatalf("failed to create model repo: %s", err.Error())
  } 

  model_commit, err := api_server.InitCommit(model_repo.Repo.Name, "master", "")
  if err != nil {
    t.Fatalf("failed to create model repo commit: %s", err.Error())
  }
  
  // delete model repo 
  if err = api_server.RemoveRepo(model_repo.Repo.Name); err != nil {
    t.Fatalf("Failed to remove model repo")
  }

  if err = api_server.RemoveRepo(test_name); err != nil {
    t.Fatalf("Failed to remove test repo")
  }

  if model_repo.Repo.Name == "" || model_commit.Commit.Id == "" {
    base.Log("Failed to generate model repo / commit")
    base.Log("[Test_GetModelRepo] Model Repo Name: ", model_repo.Repo.Name)
    base.Log("[Test_GetModelRepo] Model Commit Id : ", model_commit.Commit.Id)

    t.Fail()
  }
  
}

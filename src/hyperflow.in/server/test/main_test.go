package main_test

import (
  "os"
  "strings"
  "fmt"
  "time"
  "testing"
  "net/url"
  "io/ioutil"
  "bytes"
  "encoding/json"
  task_pkg "hyperflow.in/server/pkg/tasks"
  flow_pkg "hyperflow.in/server/pkg/flow"
  ws "hyperflow.in/server/pkg/workspace"
  client "hyperflow.in/server/pkg/test_utils/rest_client"
)

// create a rest client 
// call the server to create a new repo
// start a commit 
// send a file to server to commit 
// execute a flow 
// monitor output 
// review output 

const (
  SERVER_ADDR = "http://192.168.64.3:30001"
  API_BASE_PATH = "/"
  TEST_REPO_NAME = "test_repo0123456789"
  TEST_BRANCH_NAME ="master"
  TEST_TEMP_DIR = "/var/tmp"
  TEST_PY_FILE_NAME = "p.py"
  TEST_CMD_STRING = "python p.py"
)

func getRestClient(api_path string) client.Interface {
  base_url, _ := url.Parse(SERVER_ADDR+ API_BASE_PATH)
  client, err := client.NewRESTClient(base_url, api_path, nil)
  if err != nil {
    fmt.Println("Failed to create NewRESTClient: ", err)
  }
  return client
}

func createRepo(repoName, branchName string) (*ws.Repo, *ws.Branch, error) {
  client := getRestClient("repo")
  repo_req := client.Verb("POST")
  repo_req.Param("repoName", repoName)
  repo_req.Param("branchName", branchName)
  resp := repo_req.Do()
  _, err := resp.Raw()

  if err != nil {
    return nil, nil, fmt.Errorf("Failed while initializing repo: %s", err.Error())
  }

  return &ws.Repo{Name: repoName}, &ws.Branch{Name: "master"}, nil
}


func getRepo(repoName, branchName string) (*ws.Repo, *ws.Branch, error) {
  
  client := getRestClient("repo_attrs")
  repo_req := client.Verb("GET")
  repo_req.Param("repoName", repoName)
  repo_req.Param("branchName", branchName)
  resp := repo_req.Do()
  json_body, err := resp.Raw()

  if json_body == nil {
    return nil, nil, nil
  }

  if err != nil {
    if err_string := err.Error(); strings.Contains(err_string, "end of JSON input") {
      return nil, nil, nil
    }
    return nil, nil, fmt.Errorf("Failed while initializing repo: %s", err.Error())
  }

  repo_attrs := ws.RepoAttrs{}
  err = json.Unmarshal(json_body, &repo_attrs)
   
  return repo_attrs.Repo, repo_attrs.Branches[branchName], nil
}

func initCommit(repo *ws.Repo, branch *ws.Branch) (*ws.CommitAttrs, error) {
  client := getRestClient("commit")
  req := client.Verb("GET")
  req.Param("repoName", repo.Name)
  req.Param("branchName", branch.Name)
  req.Param("commitId", "")
  resp := req.Do()
  json_body, err := resp.Raw()

  if err != nil {
    return nil, fmt.Errorf("Failed while initializing commit: %s", err.Error())
  }

  commit_attrs := ws.CommitAttrs{}
  err = json.Unmarshal(json_body, &commit_attrs)

  if err != nil {
    return nil, err
  }

  return &commit_attrs, nil
}


func closeCommit(repo *ws.Repo, branch *ws.Branch, commit *ws.Commit) (error) {
  client := getRestClient("commit_close")
  req := client.Verb("POST")
  req.Param("repoName", repo.Name)
  req.Param("branchName", branch.Name)
  req.Param("commitId", commit.Id)

  resp := req.Do()
  _, err := resp.Raw()

  if err != nil {
    return fmt.Errorf("Failed while closing commit: %s", err.Error())
  }

  return nil
}

func createDirIfNotExist(dir string) error{
  if _, err := os.Stat(dir); os.IsNotExist(err) {
    err = os.MkdirAll(dir, 0755)
    if err != nil {
      return err
    }
  }
  return nil
}

func genSourceFile(repo_name string) (repoDir string, f *os.File, err error){
  repo_dir := TEST_TEMP_DIR + "/" + strings.Replace(repo_name, "/", "", -1)

  err = createDirIfNotExist(repo_dir)
  if err != nil {
    return "", nil, err
  }
  sep := ""
  if repo_dir[len(repo_dir)-1:] != "/" {
    sep = "/"
  }
  file_path:= repo_dir + sep + TEST_PY_FILE_NAME
  
  f, err = os.Create(file_path)
  source := []byte("print(\"hello\")\n")
  _, err = f.Write(source)

  if err != nil {
    return "", nil, err
  }

  repoDir = repo_dir
  return 
}

func genSource() ([]byte) {
  return []byte("import os \n" + 
    "print(\"Start:0\") \n" +
    "f = open(\"../saved_models/model.txt\",\"w+\") \n" + 
    "print(\"1\") \n" + 
    "f.write(\"Hello World\") \n" + 
    "print(\"2\") \n" + 
    "f.close() \n" + 
    "f1 = open(\"../out/out.txt\",\"w+\") \n" + 
    "print(\"3\") \n" +
    "f1.write(\"Hello out\") \n" + 
    "print(\"4\") \n" + 
    "f1.close() \n" +
    "print(\"5\") \n" + 
    "for z in os.listdir(\"../\"): \n" +  
    " print(z)\n" )
}

func pushCode(code []byte, fpath string, repoName string, branchName string, commitId string) (error) {
  
  client := getRestClient("file")
  r := client.Put()

  r.Param("repoName", repoName)
  r.Param("branchName", branchName)
  r.Param("commitId", commitId)
  r.Param("path", fpath)
  r.Param("size", string(len(code)))

  _ = r.SetBodyReader(ioutil.NopCloser(bytes.NewReader(code)))

  resp := r.Do()

  if resp.Error()!= nil {
    return resp.Error()
  }

  put_response := ws.PutFileResponse{}
  err := json.Unmarshal(resp.Body(), &put_response)
  if err != nil {
    return err
  }
  fmt.Println("file attrs:", put_response.FileAttrs.Object.Hash)
  fmt.Println("File written bytes: ", put_response.Written)
  return nil
}

func StartTask(repo *ws.Repo, branch *ws.Branch, commit *ws.Commit, cmdStr string) (*flow_pkg.Flow, error) {
  
  repo_msg := &ws.RepoMessage {
    Repo: repo,
    Branch: branch,
    Commit: commit,
  }

  flow_msg := flow_pkg.FlowMessage {
    Repos: []*ws.RepoMessage {
      repo_msg,
    },
    CmdStr: cmdStr,
  }
  
  json_str, _ := json.Marshal(&flow_msg) 
  
  client := getRestClient("flow")
  api := client.Verb("POST") 
  _ = api.SetBodyReader(ioutil.NopCloser(bytes.NewReader(json_str)))

  resp := api.Do()
  api_response, err := resp.Raw()  

  if err != nil {
    return nil, err
  }

  err = json.Unmarshal(api_response, &flow_msg)
  if err != nil {
    return nil, err
  }

  if flow_msg.Flow != nil {
    fmt.Println("[RunTask] Flow Id: ", flow_msg.Flow.Id)
    return flow_msg.Flow, nil
  }

  fmt.Println("[StartTask] api_response: ", string(api_response))

  return nil, fmt.Errorf("Unknown error")
}

func getTaskStatus(flow *flow_pkg.Flow) (taskStatus task_pkg.TaskStatus) {

  client := getRestClient("flow")
  url_subpath := "/" + flow.Id 

  flow_req := client.VerbSp("GET", url_subpath)
  flow_resp := flow_req.Do()

  flow_json, err:= flow_resp.Raw() 
  if err != nil {
    fmt.Println("[getTaskStatus] Warning: Failed to retrieve task status: ", err)
    return 
  }

  flow_attrs := flow_pkg.FlowAttrs{}
  _ = json.Unmarshal(flow_json, &flow_attrs)

  if task, ok := flow_attrs.Tasks[flow.Id]; ok {
    return task.Status  
  }

  return
}

func getLog(flow *flow_pkg.Flow) ([]byte, error) {
  client := getRestClient("tasks")
  url_subpath := "/" + flow.Id + "/log"

  log_req := client.VerbSp("GET", url_subpath)
  log_resp := log_req.Do()

  log_bytes, err:= log_resp.Raw()
  if err != nil {
    fmt.Println("[getLog] Failed to retrieve log: ", err)
    return nil, err
  } 
 
  return log_bytes, nil
}
 

func getCommitSize(repo *ws.Repo, branch *ws.Branch, commit *ws.Commit) (int64, error) {

  client := getRestClient("repo")
  url_subpath := "/" + url.PathEscape(repo.Name) +  "/branch/" + branch.Name + "/commit/" + commit.Id + "/attrs"

  attrs_req := client.VerbSp("GET", url_subpath) 
  attrs_resp := attrs_req.Do()

  json_body, err := attrs_resp.Raw()  
  if err != nil {
    fmt.Println("[getCommitSize] Failed the attributes request on model: ", err)
    return 0, err
  }

  model_attrs := ws.CommitAttrs{}
  if err := json.Unmarshal(json_body, &model_attrs); err != nil {
    fmt.Println("[getCommitSize] Failed to unmarshal size request response ")
    return 0, err
  }

  return model_attrs.Size, nil
}

func getSavedModel(flow *flow_pkg.Flow) (int64, error) {
  var err error

  client := getRestClient("flow")
  url_subpath := "/" + flow.Id + "/model" 

  req := client.VerbSp("GET", url_subpath)
  resp := req.Do()

  json_body, err := resp.Raw()
  if err != nil {
    fmt.Println("[getSavedModel] Failed to retrieve flow model: ", err)
    return 0, err
  }

  model := ws.RepoMessage{}
  err = json.Unmarshal(json_body, &model)

  if err != nil {
    fmt.Println("[getSavedModel] failed to unmarshal RepoMessage ")
    return 0, err
  }
    

  if model.Repo != nil {
    if model.Repo.Name != "" {
      fmt.Println("[getSavedModel] Model for flow: ", model.Repo, model.Branch, model.Commit)
      return getCommitSize(model.Repo, model.Branch, model.Commit)
    }
  }
  
  return 0, fmt.Errorf("[getSavedModel] Unknown error") 
}

func getOutputSize(f *flow_pkg.Flow) (int64, error) {
  var err error

  client := getRestClient("flow")
  url_subpath:= "/" + f.Id + "/output" 

  req := client.VerbSp("GET", url_subpath)
  resp := req.Do()

  json_body, err := resp.Raw()
  output := ws.RepoMessage{}

  err = json.Unmarshal(json_body, &output)
  if err != nil {
    fmt.Println("[getOutputSize] Failed to unmarshal repo attrs message: ", err)
    return 0, err
  }

  if output.Repo != nil {
    return getCommitSize(output.Repo, output.Branch, output.Commit)
  }

  return 0, fmt.Errorf("[getOutputSize] Unknown error") 
}

func Test_WorkerCycle(t *testing.T) {
  var repo *ws.Repo
  var branch *ws.Branch
  var err error
  var branch_name string = TEST_BRANCH_NAME
  var repo_name string = TEST_REPO_NAME
  var file_path string = TEST_PY_FILE_NAME
  var cmd_string string = TEST_CMD_STRING

  repo, branch, err= getRepo(repo_name, branch_name)
  fmt.Println("Received repo, branch:", repo, branch)
  
  if err != nil {
    fmt.Println("get_repo_error: ", err)
    t.Fatalf("get_repo_error: %s", err.Error())
  }
  
  if repo == nil || repo.Name == "" {
    fmt.Println("creating test repo... ", repo_name)
    repo, branch, err = createRepo(repo_name, branch_name)
    if err != nil {
      t.Fatalf("create_repo_error: %s", err.Error())
    }
    fmt.Println("Created repo, branch, id", repo, branch)
  } 

  commit_attrs, err := initCommit(repo, branch)
  if err != nil {
    t.Fatalf("init_commit_error: %s", err.Error())
  }
  fmt.Println("Commit Initialized: ", commit_attrs.Commit.Id)

  sample_code := genSource()
  err = pushCode(sample_code, file_path, repo.Name, branch.Name, commit_attrs.Commit.Id)
  if err != nil {
    t.Fatalf("push_code_error: %s", err)
  }

  err = closeCommit(repo, branch, commit_attrs.Commit)
  if err != nil {
    t.Fatalf("close_commit_error: %s", err)
  } 

  flw, err := StartTask(repo, branch, commit_attrs.Commit, cmd_string)
  if err != nil {
    t.Fatalf("start_task_error: %s", err)
  }
  fmt.Println("flow id: ", flw.Id)
  
  /* Check log exists */ 
  //wait for task to finish
  fmt.Println("Wait for task to finish (15s)...")
 
  time.Sleep(30 * time.Second)
 
  log, err := getLog(flw)
  if err != nil {
    //t.Fatalf("log_retrieval_error: %s", err)
  }

  if len(log) == 0 {
    fmt.Println("Expected non-zero log for this flow. Found empty.")
    //t.Fail()
  }

  var task_status task_pkg.TaskStatus
  for i := 1;  i<=3; i++ {
    task_status = getTaskStatus(flw)
    if task_status >= task_pkg.TASK_FAILED {
      break
    }
    time.Sleep(5 * time.Second)
  }

  if task_status == task_pkg.TASK_FAILED {
    fmt.Println("Task failed to complete:", task_status)
    fmt.Println(string(log)) 
    t.Fatalf("[getTaskStatus] task incomplete")
  } 

  /* Check model file exists */ 
  m_size, err := getSavedModel(flw)
  if err != nil {
    t.Fatalf("[getSavedModelSize] %s", err.Error())
  }
  if m_size == 0 {
    fmt.Println("Failed to retrieve model or found empty file")
    t.Fail()
  }

  /* Check out file exists */ 
  o_size, err := getOutputSize(flw)
  if err != nil {
    t.Fatalf("[getOutputSize] %s", err.Error())
  }
  if o_size == 0 {
    fmt.Println("Failed to retrieve output or found empty file")
    t.Fail()
  }  
}






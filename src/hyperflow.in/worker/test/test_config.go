package test_util

import( 
  "os"
  "fmt"
  "net/url"
  "encoding/json"
  
  ws "hyperflow.in/server/pkg/workspace"
  "hyperflow.in/server/pkg/base"

  "hyperflow.in/worker/utils"
  rest "hyperflow.in/worker/rest_client"
  api "hyperflow.in/worker/api_client"
  fs "hyperflow.in/worker/file_system"
  
)


var (
  DefaultServerURL = "http://192.168.64.3:30001" 
  RepoNamePrefix = "test_"
  DefaultBranchName = "master"
  TestRepoPath = "/var/tmp/work1"
  TestCodeScript = []byte("print(11)" + ";")
  DefaultFilepath = "/var/tmp/work1/train.py"
)


type TestConfig struct {
  ServerURL *url.URL
  Repo *ws.Repo
  Branch *ws.Branch
  Commit *ws.Commit
  WC *api.WorkerClient
  RepoFs *fs.RepoFs
}

func NewTestConfig() (*TestConfig, error) {
  host_url, _:= url.Parse(DefaultServerURL)

  tc := TestConfig {
    ServerURL: host_url,
  } 

  // create test repo 
  repo_name := RepoNamePrefix + utils.RandomStr(10)
  base.Println("Creating repo.. ", repo_name)
  err := tc.setRepo(repo_name)
  if err != nil {
    return nil, err
  }

  // open a new commit 
  err = tc.setCommit()
  if err != nil {
    return nil, err
  }

  err = tc.setWorkerClient()
  if err != nil {
    return nil, err
  }

  return &tc, nil
}


func (tc *TestConfig) setRepo(name string) error{

  // call API to create Repo  
  client, _ := rest.NewRESTClient(tc.ServerURL, "repo", nil)
  repo_req := client.Post() 

  repo_req.Param("repoName", name)
  resp := repo_req.Do()
  _, err := resp.Raw()

  if err != nil {
    return fmt.Errorf("failed to create standard repo: %v", err)
  }
  
  tc.Repo = &ws.Repo{
    Name: name,
  }

  // as repo API creates a default master branch
  tc.Branch = &ws.Branch{
    Name: "master",
  }

  return nil
}


func (tc *TestConfig) setBranch(repoName, name string) error {
  // call API to create branch
  return nil
}

func (tc *TestConfig) setCommit() error {
  // call APi to open a new commit
  client, _ := rest.NewRESTClient(tc.ServerURL, "commit", nil)
  repo_req := client.Get() 

  repo_req.Param("repoName", tc.Repo.Name)
  repo_req.Param("branchName", tc.Branch.Name)

  res := repo_req.Do()
  res_json, err := res.Raw()

  if err != nil {
    return fmt.Errorf("failed to create commit : %v", err)
  }

  ca := ws.CommitAttrs{}
  err  = json.Unmarshal(res_json, &ca)
  if err != nil {
    return fmt.Errorf("failed to unmarshal commit APi response, err: %v", err)
  }

  if ca.Commit != nil {
    tc.Commit = ca.Commit
  } else {
    return fmt.Errorf("failed to create commit")
  }
  return nil 
}

func (tc *TestConfig) setWorkerClient() error {
  var err error 
  tc.WC, err = api.NewWorkerClient(DefaultServerURL)
  if err != nil {
    return err
  }
  return nil
}

// create tmp repo on local system
func (tc *TestConfig) SetRepoFs() error {
  
  if err := utils.MkDirAll(TestRepoPath, os.ModePerm); err!= nil {
    return fmt.Errorf("failed to create dir %s, err: %v", TestRepoPath, err)
  }
 
  if err := utils.WriteFile(DefaultFilepath, TestCodeScript, os.ModePerm); err != nil{
    return fmt.Errorf("failed to create file %s, err: %v", DefaultFilepath, err)
  }
  
  if tc.WC == nil {
    if err := tc.setWorkerClient(); err != nil {
      return fmt.Errorf("failed to create worker client for %s, err: %v", DefaultServerURL, err)
    }
  }

  tc.RepoFs = fs.NewRepoFs(TestRepoPath, 0, tc.Repo.Name, tc.Branch.Name, tc.Commit.Id, tc.WC) 
  
  return nil
} 

func (tc *TestConfig) Destroy() error {
  // call API to delete Repo and anyhting that was created
  if tc.Repo != nil {
      client, _ := rest.NewRESTClient(tc.ServerURL, "repo", nil)
      repo_req := client.Delete() 

      repo_req.Param("repoName", tc.Repo.Name)
      resp := repo_req.Do()
      _, err := resp.Raw()

      if err != nil {
        return fmt.Errorf("failed to create standard repo: %v", err)
      }
  }

  if tc.RepoFs != nil {
    _ = os.RemoveAll(tc.RepoFs.GetWorkingDir())
  }
  return nil
}
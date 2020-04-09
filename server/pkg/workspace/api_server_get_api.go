package workspace

import( 
  "fmt" 
  "github.com/hyper-ml/hyperml/server/pkg/base"
)


func (a *apiServer) GetRepo(name string) (*Repo, error) {
  if a.q.CheckRepoExists(name) {
    return RepoRef(name), nil
  }
  return nil, errInvalidRepoName(name)
}

func (a *apiServer) GetRepoAttrs(repoName string) (*RepoAttrs, error) {
  repo_attrs, err := a.q.GetRepoAttrs(repoName)
  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return nil, err
  }

  return repo_attrs, nil
}

func (a *apiServer) GetBranchAttrs(repoName string, branchName string) (*BranchAttrs, error) {
  branch_attr, err := a.q.GetBranchAttrs(repoName, branchName)
  if err != nil {
    return nil, err
  }
  
  return branch_attr, nil
}


func (a *apiServer) GetCommitAttrs(repoName string, commitId string) (*CommitAttrs, error) {

  commit_attrs, err := a.q.GetCommitAttrsById(repoName, commitId) 
  if err !=  nil {
    return nil, err
  }

  return commit_attrs, nil
}

func (a *apiServer) GetCommitMap(repoName string, commitId string) (*FileMap, error) {
  commit_map, err := a.q.GetFileMap(repoName, commitId) 
  if err !=  nil {
    return nil, err
  } 

  return commit_map, nil
}


func (a *apiServer) GetFileAttrs(repoName string, commitId string, filePath string) (*FileAttrs, error) {

  file_attrs, err := a.q.GetFileAttrs(repoName, commitId, filePath)
  if err != nil { 
    return nil, fmt.Errorf("failed to find file details in commit %s: %v", filePath, err)
  }

  if file_attrs == nil {
    return nil, fmt.Errorf("Empty file attributes")
  }
  
  return file_attrs, nil
}

func (a *apiServer) GetLastCommitTree(repoName, branchName string) ([]*File, error){
  return nil, fmt.Errorf("unimplemented")
}

func (a *apiServer) GetCommitTree(repoName, branchName, commitId string, search_path string) (results map[string][]*File, fnerr error){
  /*
  // prepare list from map
  file_map, _ := a.GetCommitMap(repoName, commitId)
  file_paths := file_map.Filepaths(search_path)
  
  if len(file_paths) == 0 {
    return 
  } 

  dtree, _ := ListToD(file_paths)

  path_keys := dtree[search_path]
  for i :=0; i < len(path_keys); i ++ {
    file_obj := file_map.GetFile(path_keys[i])
    results = append(results, file_obj...)
  } 

  fmt.Println("results: ", results)
  return results
  // call dtree util 
  // in search path prepare list of objects and return 
  */
  return nil, fmt.Errorf("unimplemented")
}




package workspace



import(
  "fmt"
  "time"
  "strings"
  path_util "path"
  "github.com/gobwas/glob"
  "hyperview.in/server/core/utils"
  "hyperview.in/server/base"

  "hyperview.in/server/core/db"
)


type commitTxn struct {
  repoName string
  branchName string
  commitInfo *CommitInfo 
  db *db.DatabaseContext
  q *queryServer
}


func NewCommitTxn(repoName string, commitId string, db *db.DatabaseContext) (*commitTxn, error) {
  var commit_info *CommitInfo
  var err error 
  q:= NewQueryServer(db)

  if commitId != "" {
    commit_info, err = q.GetCommitInfoById(repoName, commitId)
    if (err != nil ){
      return nil, fmt.Errorf("Invalid Commit Id or Repo Name: %s", err)
    }
  }

  return &commitTxn {
    commitInfo: commit_info,
    repoName: repoName,
    branchName: "master",
    db: db,
    q: q,
  }, nil
}

func (ct *commitTxn) setCommitInfo(c *CommitInfo) {
 ct.commitInfo = c 
}

func (ct *commitTxn) IsOpenCommit() bool {
  if !ct.commitInfo.Finished.IsZero() {
    base.Log("This repo has no open commit. Please initialize commit before adding files.")
    return false
  }
  return true
}

func (ct *commitTxn) setCommitInfoByBranch() error {
  commit_info, err := ct.q.GetCommitInfoByBranch(ct.repoName, ct.branchName)
  if err != nil {
    return err
  }
  ct.commitInfo = commit_info
  return nil
}

func (ct *commitTxn) Start() (string, error) {

  var err error 
  var branch_info *BranchInfo
  var repo_info *RepoInfo
  var new_cinfo *CommitInfo
  var head_cinfo *CommitInfo

  repo_info, err = ct.q.GetRepoInfo(ct.repoName)

  if (err !=nil) {
    base.Log("InitiateCommit: Could not fetch repo with given name %s", ct.repoName)
    return "", err
  }

  // if this is first ever commit. create master branch
  if repo_info.Branch == nil {
    
    // add master branch
    branch_info, err = ct.addMasterBranch()

  } else {
    
    branch_info, err = ct.q.GetBranchInfo(ct.repoName, repo_info.Branch.Name)

    // check if there is a pending commit 
    if branch_info.Head != nil  {

      head_cinfo, err = ct.q.GetCommitInfoById(ct.repoName, branch_info.Head.Id)
      
      if head_cinfo.Finished.IsZero() {
        base.Log("There is a pending commit against this repo", head_cinfo)
        ct.setCommitInfo(head_cinfo)
        return head_cinfo.Id(), nil
      }
    }
  }
  
  if branch_info != nil {

    // add commit with current head as parent 
    new_cinfo, err = ct.addCommit(branch_info.Head)

    // update branch head with new commit  
    err = ct.scoopHead(branch_info, new_cinfo.Commit)

    if err != nil {
      //TODO : delete new commit 
      defer ct.Delete()
      return "", err
    }

  }

  if new_cinfo != nil {
    ct.setCommitInfo(new_cinfo)
    return new_cinfo.Id(), err
  }

  return "", err
}

func (ct *commitTxn) addMasterBranch() (*BranchInfo, error) {
  var err error 
  //var repo_info *RepoInfo

  repo := &Repo {
    Name: ct.repoName,
  }
  
  branch := &Branch {
    Name: "master",
    Repo: repo,
  }

  branch_info := &BranchInfo{
    Branch: branch,
    //Head: commit,
  }

  err = ct.q.InsertBranchInfo(repo.Name, ct.branchName, branch_info)

  if err != nil {
    return nil, err
  }

  //TODO: send context of error
  err = ct.q.AssignBranch(repo.Name, branch)

  if err != nil {
    return nil, err
  }

  return branch_info, err
}
  
func (ct *commitTxn) addFileMap(commit *Commit, parent *Commit) (error) {
  var err error
  var fm *FileMap = NewFileMap(commit)

  if parent != nil {
    if parent.Id != "" {
      fm, err = ct.q.GetFileMap(ct.repoName, parent.Id)
      if err != nil {
        fmt.Println("err in get file map:", err)
        return err
      }
    }
  }
  fmt.Println("insert file map:", fm)

  return ct.q.InsertFileMap(ct.repoName, commit.Id, fm)
}

func (ct *commitTxn) addCommit(parent *Commit) (*CommitInfo, error) {
  
  fmt.Println("adding file map", parent)

  var commit_info *CommitInfo
  var err error
  
  commit_id := utils.NewUUID()
  repo:= NewRepo(ct.repoName)

  commit := &Commit {
    Id: commit_id,
    Repo: repo,
  }

  commit_info = &CommitInfo {
    Commit: commit,
    Parent_commit: parent,
    Started: time.Now(),
  }

  err = ct.q.InsertCommitInfo(ct.repoName, commit_id, commit_info)
  if err != nil {
    //TODO: may be delete commit info
    return nil, err
  }
  
  if err = ct.addFileMap(commit, parent); err!= nil {
    ct.FlushCommit()
    return nil, err
  }

  return commit_info, err
}

func (ct *commitTxn) scoopHead(branchInfo *BranchInfo, commit *Commit) error {
  branch := branchInfo.Branch
  repo := branchInfo.Branch.Repo

  branchInfo.Head = commit

  err:= ct.q.UpdateBranchInfo(repo.Name, branch.Name, branchInfo)

  return err
}


func (ct *commitTxn) End() error {
  var err error 
  if (ct.commitInfo ==nil) {
    base.Log("finishCommit: Could not fetch any open commit for repo %s", ct.repoName)
    return fmt.Errorf("finishCommit: Could not fetch any open commit for repo %s", ct.repoName)
  }

  if ct.commitInfo.Finished.IsZero() {
    ct.commitInfo.Finished = time.Now()
    err = ct.q.UpdateCommitInfo(ct.repoName, ct.commitInfo.Id(), ct.commitInfo)
    return err  
  } else {
    base.Log("finishCommit: No open commit for this repo", ct.repoName)
    return fmt.Errorf("No open commit for this repo: %s", ct.repoName)
  }
  
}

func (ct *commitTxn) insertFileInfo(filePath string, object string, size int64, cs string) (*FileInfo, error) {
  var err error

  file_info := NewFileInfo(ct.commitInfo.Commit, filePath, object, size, cs)

  //TODO: get file info in return
  err = ct.q.UpsertFileInfo(ct.repoName, ct.commitInfo.Id(), filePath, file_info) 
  if err != nil {
    base.Log("Failed to update file map:", filePath, object, size)
    return nil, err 
  }

  err= ct.updateFileMap(filePath)
  if err != nil {
    base.Log("Failed to update file map:", filePath, object, size)
    return nil, err
  }

  return file_info, nil
}


func (ct *commitTxn) insertDirInfo(filePath string, size int64) (*FileInfo, error) {
  var err error
  dir_info := NewDirInfo(ct.commitInfo.Commit, filePath, size)
  err = ct.q.UpsertFileInfo(ct.repoName, ct.commitInfo.Id(), filePath, dir_info) 

  if err != nil {
    return nil, err 
  }
  err= ct.updateFileMap(filePath)

  if err != nil {
    base.Log("Failed to update file map:", filePath, size)
    return nil, err
  }
  
  return dir_info, nil 
}

func (ct *commitTxn) updateFileMap(filePath string) error {
  newfile := &File{Commit: ct.commitInfo.Commit, Path: filePath}
  return ct.q.AddFileToMap(ct.repoName, ct.commitInfo.Id(), newfile)
}

func (ct *commitTxn) AddFile(filePath string, objectName string, size int64, cs string) (*FileInfo, error) {

  if (ct.commitInfo == nil) {
    base.Log("Please initiate commit transaction with start-commit first.", ct.repoName)
    return nil, fmt.Errorf("Please initiate commit transaction with start-commit first.")
  }

  if !ct.commitInfo.Finished.IsZero() {
    return nil, fmt.Errorf("This repo has no open commit. Please initialize commit before adding files.")
  }

  if objectName == "" {
    return ct.insertDirInfo(filePath, size)
  }

  return ct.insertFileInfo(filePath, objectName, size, cs)
}

func (ct *commitTxn) AddDir(filePath string, size int64) (*FileInfo, error) {

  // TODO: get the latest commit info to avoid concurrency issues
  if !ct.commitInfo.Finished.IsZero() {
    return nil, fmt.Errorf("This repo has no open commit. Please initialize commit before adding files.")
  }
  
  return ct.insertDirInfo(filePath, size)
}

func (ct *commitTxn) FlushCommit() error{
  //delete commit and the file map
  return ct.Delete()
}


func (ct *commitTxn) Delete() error {
  // delete commit 
  var err error
  var branch_info *BranchInfo

  if ct.commitInfo == nil {
    if err = ct.setCommitInfoByBranch(); err != nil {
      return err
    }
  }

  if !ct.IsOpenCommit() {
    return fmt.Errorf("This repo has no open commit to flush")
  } 

  if ct.commitInfo.Parent_commit != nil {
    branch_info, err = ct.q.GetBranchInfo(ct.repoName, ct.branchName)
    if err != nil {
      base.Log("Invalid repo or branch name:", ct.repoName, ct.branchName)
      return err
    }
    if err:= ct.scoopHead(branch_info, ct.commitInfo.Parent_commit); err!= nil {
      base.Log("Unable to scoop branch head to parent:", ct.commitInfo.Parent_commit.Id)
      return err
    }

  }

  return ct.q.DeleteCommitInfo(ct.repoName, ct.commitInfo.Id())
}


// list files and sub directories given a directory path
//
func (ct *commitTxn) lsDir(list map[string]*File, prefix string) (map[string]*FileInfo, error) {

  result:= make(map[string]*FileInfo)
  
  /* Commented as client should send root path
  if prefix == "" {
    prefix = "/"
  }*/
  
  fmt.Println("prefix:", prefix)

  if prefix !="" && prefix[len(prefix)-1:] == "*"   {
    prefix  = prefix[:len(prefix)-1]
  }
  
  glob_pattern := prefix

  if glob_pattern[len(glob_pattern)-1:] != "/"   {
    glob_pattern  = glob_pattern + "/"
    prefix = prefix + "/"
  }

  g := glob.MustCompile(glob_pattern + "*")

  // / root doesnt work

  for path, file := range list {
    if g.Match(path) { 

      var path_splits []string
      var path_woprefix string
      
      //path without prefix 
      if prefix != "/" { 
        path_woprefix = strings.Replace(path, prefix, "", -1)
      } else {
        path_woprefix = path[1:] 
      }
 
      if len(path_woprefix) > 0 {
        path_splits = strings.SplitAfter(path_woprefix, "/")
      }

      if len(path_splits) >0 { 
        path_woslash := strings.Replace(path_splits[0], "/","", -1)

        if len(path_woslash) > 0 {

          if path_woslash == path_util.Base(file.Path) {
            file_info, err := ct.q.GetFileInfo(file.Commit.Repo.Name, file.Commit.Id, path)
            
            if err == nil {
              result[path_woslash] = file_info
            } else {
              base.Log("something wrong. File Info missing for file: %s %s %s", file.Commit.Repo.Name, file.Commit.Id, path)
            } 

          } else {
            fmt.Println("creating directory")
            dir_info := NewDirInfo(file.Commit, path_woslash, 0)
            result[path_woslash] = dir_info
          } 

          // for directory create a new file info object and respond 
        }
      }
    }
      
  }
 
  return result, nil
}

// list directory path
func (ct *commitTxn) ListDir(dirPath string) (map[string]*FileInfo, error) {
  
  if ct.commitInfo == nil {
    return nil, fmt.Errorf("Missing Commit Info. Please start commit transaction with Id or start a new commit.")
  }

  fm, err := ct.q.GetFileMap(ct.repoName, ct.commitInfo.Id())
  if err != nil {
    return nil, fmt.Errorf("Commit has not files or dirs to list")
  }

  return ct.lsDir(fm.Entries, dirPath)
}

// handle full path or just look at base path?

func (ct *commitTxn) LookupFile(fpath string) (*FileInfo, error) {

  if ct.commitInfo == nil {
    return nil, fmt.Errorf("[commitTxn.LookupFile] Missing Commit Info. Please start commit transaction with Id or start a new commit.")
  }

  fm, err := ct.q.GetFileMap(ct.repoName, ct.commitInfo.Id())
  if err != nil || fm == nil {
    return nil, fmt.Errorf("[commitTxn.LookupFile] Commit has not files or dirs to list")
  }
  
  base.Debug("[commitTxn.LookupFile] fpath parameter - ", fpath)

  if fe := fm.Entries[fpath]; fe != nil {
    base.Debug("[commitTxn.LookupFile] found file in entries - ", fpath, ct.commitInfo.Id())

    file_info, err := ct.q.GetFileInfo(ct.repoName, fe.Commit.Id, fe.Path)
    

    if err == nil {
      base.Debug("[commitTxn.LookupFile] File Info of file found", file_info.File.Path)
      return file_info, nil
    } else if !base.IsErrFileNotFound(err) {
      base.Debug("[commitTxn.LookupFile] Unknown Error while looking for file. ", err)
      return nil, err
    }
  }
  
  // check if input name is a directory

  glob_pattern := fpath

  g := glob.MustCompile(glob_pattern + "/*") 

  for p, _ := range fm.Entries {   
    if g.Match(p) { 
      dir_info := NewDirInfo(ct.commitInfo.Commit, fpath, 0)
      return dir_info, nil
    }
  }

  return nil, &base.ErrFileNotFound{CommitId: ct.commitInfo.Id(), RepoName: ct.repoName, Fpath: fpath}
}






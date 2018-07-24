package file_system

import ( 
  "bazil.org/fuse"
 "bazil.org/fuse/fs" 

  ws "hyperview.in/server/core/workspace" 
 
  "hyperview.in/worker"
)

func getRepoHead(wc *worker.WorkerClient, repoName string, branchName string) (*ws.Commit, error) {
  //var repo_info ws.RepoInfo

  branch_info, err := wc.FetchBranchInfo(repoName, branchName)

  if err != nil {
    return nil, err
  } 
  return branch_info.Head, nil
}
 

// mount repo on local path
//
func mount(repoName, repoRemotePath, repoLocalPath string) error {
  conn, err := fuse.Mount(repoLocalPath)
  if err != nil {
    return err
  }

  defer conn.Close()
  
  wc, err := worker.NewWorkerClient()

  if err != nil {
    return err
  }

  head, err := getRepoHead(wc, repoName, "master")
  if err != nil {
    return err
  } 

  filesys := &FS {
    wc: wc,
    repoName: repoName,
    branchName: "master",
    head: head,
    iNodeRegistry: make(map[string]uint64),
  }

  if err := fs.Serve(conn, filesys); err != nil {
    return err
  }

  <- conn.Ready
  if err := conn.MountError; err != nil {
    return err
  }

  return nil
}

//TODO
func Unmount(dir string) {
  fuse.Unmount(dir)
}

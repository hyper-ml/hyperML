package fuse

import ( 
  "fmt"
  "bazil.org/fuse"
  "bazil.org/fuse/fs" 

  "github.com/hyper-ml/hyperml/server/pkg/base"
  ws "github.com/hyper-ml/hyperml/server/pkg/workspace" 
 
  api_client "github.com/hyper-ml/hyperml/worker/api_client"
)

func getRepoHead(wc *api_client.WorkerClient, repoName string, branchName string) (*ws.Commit, error) {
  //var repo_attrs ws.RepoAttrs

  branch_attr, err := wc.FetchBranchAttrs(repoName, branchName)

  if err != nil {
    return nil, err
  } 
  return branch_attr.Head, nil
}
 

// mount repo on local path
//
func mount(repoName string, repoRemotePath string, repoLocalPath  string, wc *api_client.WorkerClient) (*Mounter, error) {
  /*conn, err := fuse.Mount(repoLocalPath)
  if err != nil {
    return nil, err
  }

  defer conn.Close()*/
    
  head, err := getRepoHead(wc, repoName, "master")
  if err != nil {
    return nil, err
  } 
  fmt.Println("head repo: ", head)

  filesys := &FS {
    wc: wc,
    repoName: repoName,
    branchName: "master",
    head: head,
    iNodeRegistry: make(map[string]uint64),
  }
  conf := &fs.Config{}

  mnt, err := Mounted(repoLocalPath, filesys, conf)
  if err != nil {
    base.Log("[mount_func.mount] Failed to mount directory: ", repoLocalPath, err)
  }
 
  return mnt, nil
/*
  if err := fs.Serve(conn, filesys); err != nil {
    return err
  }

  <- conn.Ready
  if err := conn.MountError; err != nil {
    return err
  }
*/ 
}

//TODO
func Unmount(dir string) {
  fuse.Unmount(dir)
}

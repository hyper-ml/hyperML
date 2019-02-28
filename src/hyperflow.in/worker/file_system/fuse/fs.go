package fuse


import (
  "fmt"
  "sync"
  "bazil.org/fuse/fs"
  //"hyperflow.in/server/pkg/base"

  ws "hyperflow.in/server/pkg/workspace"

  api_client "hyperflow.in/worker/api_client"
)

type FS struct {
  wc *api_client.WorkerClient
  repoName string
  head *ws.Commit
  branchName string 
  iNodeRegistry map[string]uint64
  lockRegistry   sync.RWMutex
  Mounter *Mounter
}
 

var _ = fs.FS(&FS{})

func (f *FS) Root() (n fs.Node, ret error) {
  
  n = &Dir {
    fs: f,
    Node: Node {
      WsFile: ws.File {
        Commit: f.head,
        Path: "/",
      },
    RepoName: f.repoName,
    HeadCommitId: f.head.Id,
    Write: true,
    },
  }
  return n, nil
}

 
func (f *FS) iNodeNumber(file *ws.File) uint64 {
  
  f.lockRegistry.RLock()
  file_key:= getFileKey(file)

  iNode, ok := f.iNodeRegistry[file_key]
  f.lockRegistry.RUnlock()

  if ok {
    return iNode
  }

  f.lockRegistry.Lock()
  defer f.lockRegistry.Unlock()

  iNode, ok = f.iNodeRegistry[file_key]

  if ok {
    return iNode
  }

  newNodeNum := uint64(len(f.iNodeRegistry))
  f.iNodeRegistry[file_key] = newNodeNum
  return newNodeNum
}

func getFileKey(file *ws.File) string {    
  return fmt.Sprintf("%s/%s/%s", file.Commit.Repo.Name, file.Commit.Id, file.Path)
}
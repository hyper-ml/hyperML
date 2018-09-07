package workspace

import (
  "io"

  "hyperview.in/server/base"
  
  "hyperview.in/server/core/storage"
  dbpkg "hyperview.in/server/core/db"
)

// Virtual File Server
// what: virtual file server on commit files
// why: keep file related activities separate from API Server

type VfsServer struct {
  db *dbpkg.DatabaseContext
  objectAPI *ObjWrapper

}

func NewVfsServer(d *dbpkg.DatabaseContext, oapi storage.ObjectAPIServer) *VfsServer{
  return &VfsServer {
    db: d,
    objectAPI: &ObjWrapper{
      api: oapi,
    }, 
  }
}

// List files in a given commit from a directory
func (vfs *VfsServer) ListDir(repoName string, branchName string, commitId string, path string) (map[string]*FileAttrs, error) {
  var err error
  
  ctxn, err := NewCommitTxn(repoName, branchName, commitId, vfs.db)
  if err != nil {
    base.Log("Failed to start a commit txn with given params: %s %s", repoName, commitId)
  }

  if commitId == "" {
    base.Debug("No Commit passed. Using open commit: %s", commitId)
    
    c, err := ctxn.Start()
    base.Log("c", c)
    if err != nil {
      base.Log("Failed to start commit with id: %s %s %s", repoName, commitId, err)
      return nil, err
    }
  }

  files, err := ctxn.ListDir(path)
  if err != nil {
    base.Log("Failed to retrieve files in the given directory:", err)
    return nil, err
  }

  return files, err
}


func (vfs *VfsServer) Lookup(repoName string, branchName string, commitId string, p string) (*FileAttrs, error) {
  var f_info *FileAttrs
  var err error

  ctxn, err := NewCommitTxn(repoName, branchName, commitId, vfs.db)
  if err != nil {
    base.Log("[VfsServer.Lookup] Failed to start a commit txn with given params: %s %s", repoName, commitId)
    return nil, err
  }

  if commitId == "" {
    base.Debug("[VfsServer.Lookup] No Commit passed. Using open commit: %s", commitId)
    
    commit_id, err := ctxn.Start()
    if err != nil {
      base.Log("[VfsServer.Lookup] Failed to start commit with id: %s %s %s", repoName, commitId, err)
      return nil, err
    }
    _ = commit_id
  }

  f_info, err = ctxn.LookupFile(p)

  if err != nil {
    base.Log("[VfsServer.Lookup] Failed to retrieve file or directory:", err)
    return nil, err
  }

  return f_info, nil

}


func (vfs *VfsServer) PutFileByHash(hash string, reader io.Reader) (int64, error) {
  _, _, size, err :=  vfs.objectAPI.AppendObject(hash, reader) 
  return size, err    
}





















package workspace

/* -- commented as fuse is unlikely to be used in future versions 

import (
  "io"
  
  "hyperflow.in/server/pkg/storage"
  dbpkg "hyperflow.in/server/pkg/db"
)

type VfsServer struct {
  db dbpkg.DatabaseContext
  objects *storage.ObjectAPIServer

}

func NewVfsServer(d dbpkg.DatabaseContext, oapi storage.ObjectAPIServer) *VfsServer{
  return &VfsServer {
    db: d,
    objects: &oapi, 
  }
}

// List files in a given commit from a directory
func (vfs *VfsServer) ListDir(repoName string, branchName string, commitId string, path string) (map[string]*FileAttrs, error) {
  var err error
 
  ctxn, err := NewCommitTxn(repoName, branchName, commitId, vfs.db)
  if err != nil {
    return nil, err
  }

  if ctxn.GetCommitAttrs() == nil {  
    c, err := ctxn.Start()
    if err != nil {
      return nil, err
    }
  }

  files, err := ctxn.ListDir(path)
  
  if err != nil {
    return nil, err
  }

  return files, err
}


func (vfs *VfsServer) Lookup(repoName string, branchName string, commitId string, p string) (*FileAttrs, error) {
  var f_info *FileAttrs
  var err error

  ctxn, err := NewCommitTxn(repoName, branchName, commitId, vfs.db)
  if err != nil {
    return nil, err
  }

  if commitId == "" {
    
    commit_id, err := ctxn.Start()
    if err != nil {
      return nil, err
    }
    _ = commit_id
  }

  f_info, err = ctxn.LookupFile(p)

  if err != nil {
    return nil, err
  }

  return f_info, nil

}


func (vfs *VfsServer) PutFileByHash(hash string, reader io.Reader) (int64, error) {
  _, _, size, err :=  vfs.objects.SaveObject(hash, reader, false) 
  return size, err    
}*/





















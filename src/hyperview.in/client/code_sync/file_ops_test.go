package code_sync

import(
  "fmt"
  "testing" 
  "net/url"
  "hyperview.in/client/fs"
)


func Test_GetByUrl(t *testing.T) { 
  baseUrl, _ := url.Parse("http://localhost:8888")

  // load local FS 
  lfs := fs.NewFS("/Users/apple/MyProjects/stash", 0)
  
  f:= NewFileOp(nil, baseUrl, "object", lfs)
  f = f.RemoteURLParams("commitId", "b2874bdc9cfd4d0cacd8b8091e90fa93")
  f = f.RemoteURLParams("repoName", "test_repo")
  f = f.RemoteURLParams("filePath", "file")
  //fmt.Println("f:", f.RemoteURL())
  r, err := f.GetFileByUrl("/objects/file")
  fmt.Println("r:",r,err )
  if err != nil {
    t.Fatalf("Failed to get file: %s", err)
  }
}
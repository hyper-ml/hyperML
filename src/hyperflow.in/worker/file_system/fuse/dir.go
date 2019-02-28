package fuse

import (
  "os" 
  "time"
  "fmt"
  "strings"
  path_pkg "path"
  "golang.org/x/net/context"

  "bazil.org/fuse"
  "bazil.org/fuse/fs"
  "hyperflow.in/server/pkg/base"
  
  ws "hyperflow.in/server/pkg/workspace"

)

type Node struct {
  WsFile ws.File
  RepoName string
  HeadCommitId string
  Write bool
}

type Dir struct {
  fs *FS
  Node
}

var _ = fs.Node(&Dir{})


func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) (funcError error) {
  // TODO : log outcome 

  a.Valid = time.Nanosecond
  if d.Write {
    a.Mode = os.ModeDir | 0775
  } else {
    a.Mode = os.ModeDir | 0555
  }
  
  //TODO: a.Inode = d.fs.iNodeNumber(&d.WsFile)
  //TODO : add a.mTime
  
  return nil
}



func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
  base.Debug("[dir.ReadDirAll] Dir Path: ", d.WsFile.Path)

  var res []fuse.Dirent
  
  if d.WsFile.Path == "" {
    return res, nil
  }
  
  fInfo_map, err := d.fs.wc.ListDir(d.RepoName, d.HeadCommitId, d.WsFile.Path)
 
  if err != nil {
    return res, err
  }
  
   for fpath, fileAttr := range fInfo_map {
 
    if fileAttr.FileType == "DIR" {

       de := fuse.Dirent{
        Name: fpath,
        Type: fuse.DT_Dir,
      }
      res = append(res, de)
    } else {
       de := fuse.Dirent{
        Name: fpath,
        Type: fuse.DT_File,
      }
      res = append(res, de)
    }

  }

  //base.Debug("ReadDirAll result:", res)
  return res, nil
}



//TODO: invalid directory is also going through
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
  return d.lookupFile(ctx, name)
}

func stripAndJoinPath(parent_dir string, search_name string) string {
  

  if search_name == "" {
    return ""
  }

  switch {
  case parent_dir == "/":
    return search_name 
  case string(parent_dir[len(parent_dir)-1:]) != "/":
    return parent_dir + "/" + search_name  
  default:
    return parent_dir + search_name
  }

 /* if  name != "" &&  string(p[len(p)-1:]) != "/" {
    n = "/" + name  
  } else if name != "" {
    n = name
  }  

  return p  + n */

  return ""
}

func (d *Dir) lookupFile(ctx context.Context, name string) (fs.Node, error) {
  
  var dir_node *Dir

  if name[0] == '.' {
    return nil, fuse.ENOENT 
  }

  var f_info *ws.FileAttrs
  var err error

  if d.WsFile.Commit == nil {
   return nil, fuse.ENOENT 
  }

  parent_path := d.WsFile.Path
  

  f_info, err = d.fs.wc.LookupFile(d.RepoName, d.HeadCommitId, stripAndJoinPath(parent_path, name))
  
  if err != nil {

    if strings.Contains(err.Error(), "not found") {
      base.Log("[Dir.lookupFile] File or directory not found on server. ", err)
      return nil, fuse.ENOENT 
      } else {
        base.Log("[Dir.lookupFile] Error fetching file info from server.", err)
        return nil, fuse.EIO
    }
  }  
  
  if d.Write {
    f_info.SizeBytes = 0
  }

  dir_node = d.copy()
  dir_node.WsFile.Path  = f_info.File.Path 
  dir_node.WsFile.Commit = f_info.File.Commit

  switch f_info.FileType {

  case "FILE":
    return &file{
      Dir: *dir_node,
      Size: int64(f_info.SizeBytes),
    }, nil

  case "DIR":
    return dir_node, nil

  default: 
    return nil, fmt.Errorf("unrecognized file type")
  }
}

var _ = fs.NodeCreater(&Dir{})

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (result fs.Node, _ fs.Handle, createError error) {
  base.Log("[d.Create]  req.Name",  req.Name)
  base.Log("[d.Create]  d.WsFile.Path",  d.WsFile.Path)

  if d.HeadCommitId == "" {
    return nil, 0, fuse.EPERM
  }

  dn := d.copy()
  dn.WsFile.Path  = path_pkg.Join(dn.WsFile.Path, req.Name)
  newf := &file{
    Dir: *dn,
    Size: 0,
  }

  if err := newf.createEmpty(); err != nil {
    // check if commit is open
    // raise rror
    return nil, 0, err
  }
  resp.Flags |= fuse.OpenDirectIO | fuse.OpenNonSeekable
  fh := newf.newFileHandle(0)

  return newf, fh, nil
}

func (d *Dir) copy() *Dir {
  base.Log("directory in copy: ", d.WsFile.Path)

  return &Dir {
    fs: d.fs,
    Node: Node {
      WsFile: ws.File {
        Commit: &ws.Commit {
          Repo: &ws.Repo {
            Name: d.WsFile.Commit.Repo.Name,
          },
          Id: d.WsFile.Commit.Id,
        }, 
        Path: d.WsFile.Path,
      },
      HeadCommitId: d.HeadCommitId,
      RepoName: d.RepoName, 
      Write: d.Write,
    },
  }
}


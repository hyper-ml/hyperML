package fuse

import ( 
  "io" 
  "sync"
  "bytes"
  "fmt"
  "golang.org/x/net/context"
  
  "bazil.org/fuse"
  "bazil.org/fuse/fs"

  "hyperflow.in/server/pkg/base"

)
type file struct {
  Dir 
  Size int64

  lock sync.Mutex
  //count of writers
  fileHandles []*fileHandle
}



func (f *file) Attr(ctx context.Context, a *fuse.Attr) (funcErr error) {
  base.Info("[file.Attr] path: ", f.WsFile.Path)
  file_attrs, err := f.fs.wc.FetchFileAttrs(f.RepoName, f.WsFile.Commit.Id, f.WsFile.Path)

  if err != nil {
    return err
  }
 
  a.Size = uint64(file_attrs.SizeBytes)
 
  a.Mode =  0775 
  a.Inode = f.fs.iNodeNumber(&f.WsFile)
  return nil
}

func (f *file) newFileHandle(pos int ) *fileHandle {
  base.Info("[file.newFileHandle] path: ", f.WsFile.Path)

  var w io.WriteCloser
  var err error 

  f.lock.Lock()
  defer f.lock.Unlock()  

  // send head commit instead of file commit 
  // to avoid overwriting old commits
  w, err = f.fs.wc.PutFileWriter(f.RepoName, f.HeadCommitId, f.WsFile.Path) 

   if err != nil {
    base.Log("Unable to create a server writer for file:", f.WsFile.Path)
  }
  fh := &fileHandle{f: f, pos: pos, writer: w}

  f.fileHandles = append(f.fileHandles, fh)
  return fh
}

func (f *file) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (_ fs.Handle, retErr error) {
  base.Info("[file.Open] path: ", f.WsFile.Path)

  resp.Flags |= fuse.OpenNonSeekable
  file_attrs, err := f.fs.wc.FetchFileAttrs(f.WsFile.Commit.Repo.Name, f.WsFile.Commit.Id, f.WsFile.Path)
  if err!= nil {
    base.Log("Open: Unable to fetch file info from server:", err)
    return nil, err
  }

  h := f.newFileHandle(int(file_attrs.SizeBytes))
  return h, nil
}


func (f *file) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
  base.Info("[file.Fsync] path: ", f.WsFile.Path)

  f.lock.Lock()
  defer f.lock.Unlock()
  for _, fh := range f.fileHandles {
    if fh.writer != nil {
      w := fh.writer
      fh.writer = nil
      if err := w.Close(); err != nil {
        return err
      }
    }
  }
  return nil
}

func (f *file) createEmpty() error {
  
  base.Log("[f.createEmpty] In createEmpty :", f.WsFile.Path)

  w, err := f.fs.wc.PutFileWriter(f.RepoName, 
          f.HeadCommitId, 
          f.WsFile.Path,
        ) 

  if err != nil {
    base.Log("Unable to create a server writer for file:", f.WsFile.Path)
    return err
  }

  if _, err = w.Write([]byte("")); err != nil {
    base.Log("[f.createEmpty] Failed to write to an empty file on server")
    return err
  }

  return w.Close()
}

type fileHandle struct {
  f *file
  writer io.WriteCloser
  c int
  // current position of pointer. size read so far 
  pos int
  lock sync.Mutex
}



// write req.Data at req.offset
// return resp.Size with bytes written

func (fh *fileHandle) Write(ctx context.Context, r *fuse.WriteRequest, resp *fuse.WriteResponse) error {
  base.Info("[fileHandle.Write] ")
  // synchronize writes 
  fh.lock.Lock()
  defer fh.lock.Unlock()

  offset := fh.pos - int(r.Offset)  

  if offset < 0 {
    return fmt.Errorf("Found gap in bytes received. Something is wrong")
  }

  if offset > len(r.Data) {
    return nil
  }

  sent_bytes, err := fh.writer.Write(r.Data[offset:])
  if err != nil {
    base.Log("Encounted error while writing file (%s) to server: %s", fh.f.WsFile.Path, err)
    return err
  }
  
  defer fh.writer.Close()

  fh.pos = fh.pos + sent_bytes
  resp.Size = offset + sent_bytes

  // update file size in case the request offset + bytes 
  // from this request have crossed file size

  if fh.f.Size < r.Offset + int64(sent_bytes) {
    fh.f.Size  = r.Offset + int64(sent_bytes)
  } 
    
  return nil 
}


func (fh *fileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
  var buf bytes.Buffer 
  base.Info("[fileHandle.Read] fh.f.WsFile.Path:", fh.f.WsFile.Path)
  err := fh.f.fs.wc.GetFileObject(
      fh.f.WsFile.Commit.Repo.Name,
      fh.f.WsFile.Commit.Id,
      fh.f.WsFile.Path,
      req.Offset,
      int64(req.Size),
      &buf)

  //TODO: handle unexpected EOF
  if err != nil && err != io.EOF && err.Error() != "unexpected EOF" {
    base.Log("Failed to Get object from fileHandle.Read(): ", err)
    return err
  }

  resp.Data = buf.Bytes()

  return nil
}


/* 

func (fh *fileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
  return nil
}*/ 
 
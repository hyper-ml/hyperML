package fuse

import (
  //"io/ioutil"
  "errors"
  "log"
  "os"
  "time"

  "fmt"
  "bazil.org/fuse"
  "bazil.org/fuse/fs" 

)
 
type Mounter struct {
  // Dir is the temporary directory where the filesystem is mounted.
  Dir string

  Conn   *fuse.Conn
  Server *fs.Server

  // Error will receive the return value of Serve.
  Error <-chan error

  done   <-chan struct{}
  closed bool
}
 

func Mounted(dir string, filesys fs.FS, conf *fs.Config, options ...fuse.MountOption) (*Mounter, error) {  
  //dir, err := dir

  //ioutil.TempDir("", dir)
  //if err != nil {
  //  return nil, err
  //}

  c, err := fuse.Mount(dir, options...)
  if err != nil {
    return nil, err
  }

  server := fs.New(c, conf)
  done := make(chan struct{})
  serveErr := make(chan error, 1)

  mnt := &Mounter{
    Dir:    dir,
    Conn:   c,
    Server: server,
    Error:  serveErr,
    done:   done,
  }
  
  go func() {
    defer close(done)
    serveErr <- server.Serve(filesys)
  }()

  select {
  case <-mnt.Conn.Ready:
    if err := mnt.Conn.MountError; err != nil {
      return nil, err
    }
    return mnt, nil
  case err = <-mnt.Error:
    // Serve quit early
    if err != nil {
      return nil, err
    }
    return nil, errors.New("Serve exited early")
  }
}

func (mnt *Mounter) Close() {
  fmt.Println("Mounter close")
  if mnt.closed {
    return
  }
  mnt.closed = true
  for tries := 0; tries < 1000; tries++ {
    err := fuse.Unmount(mnt.Dir)
    if err != nil {
      // TODO do more than log?
      log.Printf("unmount error: %v", err)
      time.Sleep(10 * time.Millisecond)
      continue
    }
    break
  }
  <-mnt.done
  mnt.Conn.Close()
  os.Remove(mnt.Dir)
}



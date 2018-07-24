package code_sync

import (  
  "net/url"  
  "io"
  "hyperview.in/client/utils"
  "hyperview.in/client/fs"  
)


// file driver 
type FileOp struct {

  //http related flags 
  r *remoteRequest 

  localfs fs.FS 

  //TODO: add size and other controls on push/pull operations
} 

 
func NewFileOp(client HTTPClient, baseURL *url.URL, subPath string, localfs fs.FS) *FileOp{
  return &FileOp{
    r: &remoteRequest {
      client: client,
      baseURL: baseURL,
      subPath: subPath,
    }, 
    localfs: localfs,
  }
}

func (f *FileOp) RemoteURLParams(paramName, v string ) *FileOp {
  if f.r.err != nil {
    return f
  }
  f.r = f.r.setParam(paramName, v)
  return f
}

func (f *FileOp) RemoteURL() *url.URL {
  return f.r.URL()
}

func (f *FileOp) GetFileByUrl(fileName string) (int64, error) {
  var r_bytes int64
  var w_bytes int64

  // reader remote object  
  r, err := f.r.ReadResponse()
  
  if err != nil {
    return 0, err
  }

  defer r.Close()

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  // TODO: retry 

  for {
    n, err := r.Read(buf)
    if n == 0 && err != nil {
      if err == io.EOF {
        return r_bytes, nil
      }
      return r_bytes, err
    }

    w_bytes, err = f.localfs.MakeFile(fileName, func(w io.Writer) error {
      _, err:= w.Write(buf[:n])
      return err
    }) 
    if err != nil {
      return w_bytes, err
    }
  }

  return w_bytes, nil 
}







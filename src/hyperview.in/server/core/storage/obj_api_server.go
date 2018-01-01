package storage

import (
  "fmt"
  "io"
  "strings"
  "io/ioutil"  

  "path/filepath"
  "golang.org/x/net/context"

  "github.com/golang/groupcache"
  "hyperview.in/server/core/utils"
  "hyperview.in/server/base"
)

var (
  splitSize = int64(16 * 1024 * 1024)
)


type ObjectAPIServer interface {
  BasePath() string
  GetObjectPath(subpath string) string

  Reader(path string, offset int64, size int64) (io.ReadCloser,error)
  ReadSeeker(path string, offset int64, size int64) (io.ReadSeeker, error)
  Writer(path string) (io.WriteCloser, error) 
  
  CreateObject(parentDir string, fileReader io.Reader, inchunks bool) (string, string, int64, error)
  SaveObject(objName string, parentDir string, fileReader io.Reader, inchunks bool) (obj_path string, chksm string, written int64, retErr error) 

  UpdateObject(parentDir string, objectHash string, fileReader io.Reader, inchunks bool) (string, string, int64, error)
  GetObject(name string, offset int64, size int64) ([]byte, int, error) 
  CheckObject(name string) bool
  DeleteObject(path string) error
  ReadObject(objPath string, offset int64, size int64, objWriter io.Writer ) (int64, error)
}
 
type objAPIServer struct {
	//log.Logger
	dir string
	storageClient Client

  // cache structures
  objCache *groupcache.Group
  objCacheBytes int64
}



// TODO: cache small files. 
// Google storage is fast enough so this need can be parked for now
func newObjectAPIServer(dir string, client Client, cacheBytes int64) (*objAPIServer, error){
	 
  o := &objAPIServer{
    dir: dir,
    storageClient: client,
  } 
  return o, nil
}

func newGoogleStorageAPIServer(dir string, cacheBytes int64) (*objAPIServer, error) {
	bucket := base.GetEnv("GOOGLE_STORAGE_BUCKET")
  
  if bucket == "" {
    return nil, fmt.Errorf("Set GOOGLE_STORAGE_BUCKET variable to use google storage option")
  }

	storageClient, err := NewGoogleClient(context.Background(), bucket)
	
  if err != nil {
    fmt.Println("Error occured creating google client:", err)
		return nil, err
	}

	return newObjectAPIServer(dir, storageClient, cacheBytes)
}

func newGoogleBucketAPIServer(bucket string, dir string, cacheBytes int64) (*objAPIServer, error) {
  
  if bucket == "" {
    return nil, fmt.Errorf("Set GOOGLE_STORAGE_BUCKET variable to use google storage option")
  }

  storageClient, err := NewGoogleClient(context.Background(), bucket)
  
  if err != nil {
    fmt.Println("Error occured creating google client:", err)
    return nil, err
  }

  return newObjectAPIServer(dir, storageClient, cacheBytes)
}

func (s *objAPIServer) BasePath() string {
  return s.dir
}

func (s *objAPIServer) GetObjectPath(subpath string) string {
  return filepath.Join(s.dir, subpath)
}


func (s *objAPIServer) objDir(path string) string {
  return filepath.Join(s.dir, path)
}

func (s *objAPIServer) objPath(path string, file_name string) string {
  return filepath.Join(s.objDir(path), file_name)
}

func (s *objAPIServer) Writer(path string) (io.WriteCloser, error) {
  return s.storageClient.Writer(path)
}

func (s *objAPIServer) Reader(path string, offset int64, size int64) (io.ReadCloser,error) {
  reader, err := s.storageClient.Reader(path, offset , size)
  
  if (err != nil) {
    //TODO: temporary hack for now to detect end of file
    if strings.Contains(err.Error(), "InvalidRange") {
       return nil, io.EOF
    } 
    return nil, err
  }

  return reader, err
}

func  (s *objAPIServer) ReadSeeker(fpath string, offset int64, size int64) (io.ReadSeeker, error) {
  var object_size int64
  base.Info("[objApiServer.ReadSeeker] file path: ", fpath)

  reader, err := s.Reader(fpath, offset, size)
  
  if err != nil {
    return nil, err
  }

  if size != 0 {
    object_size = size 
  } else {
    object_size = s.GetObjectSize(fpath)
  }

  return &getObjReadSeeker{
    reader: reader,
    obj_path: fpath,
    offset: offset,
    size: object_size,
    s: *s,
  }, nil
}
func (s *objAPIServer) CreateObject(parentDir string, fileReader io.Reader, inchunks bool) (obj_path string, chksm string, written int64, retErr error) {
  obj_hash := utils.NewUUID()
  base.Debug("[objApiServer.CreateObject] obj_hash: ", obj_hash)
  return s.SaveObject(obj_hash, parentDir, fileReader, inchunks)
}
  
// TODO: add error channel and submit a go func to write object
// TODO: add cancel with context
//
func (s *objAPIServer) SaveObject(objName string, parentDir string, fileReader io.Reader, inchunks bool) (obj_path string, chksm string, written int64, retErr error) {
	base.Info("[objApiServer.SaveObject] objName, parentDir: ", objName, parentDir )
  var err error 

  cksum_hash := utils.NewHash()

  obj_path = s.objPath(parentDir, objName)
  
  r := io.TeeReader(fileReader, cksum_hash) 
  
	wc, err := s.storageClient.Writer(obj_path) 

  defer func() {

  }()
  
  if err != nil {
    return obj_path,"", 0, err
  }

  if inchunks {
    written, err = io.CopyN(wc, r, splitSize)
    _, _ = written, err
  } else {
    // create and use own buffer queue to limit memory usage
    buffer := utils.GetBuffer()
    defer utils.PutBuffer(buffer)
    written, err = io.CopyBuffer(wc, r, buffer)
    _, _ = written, err
  }
  
  if err != nil {
    if err != io.EOF {
      // delete half created object
      s.storageClient.Delete(obj_path)
      base.Log("[objAPIServer.CreateObject] Failed to write object:", err)
      retErr = err
      return
    }
    if err == io.EOF && written == 0 {
      return
    }
  }

  retErr = wc.Close() 
  if retErr != nil {
    written = 0 
    chksm = ""
    return
  } 

  // calculate checksum 
  chksm = utils.HexToString(cksum_hash.Sum(nil))
  retErr = err
  return 
}

func (s *objAPIServer) UpdateObject(parentDir string, objectHash string, fileReader io.Reader, inchunks bool) (objPath string, writesum string, written int64, retErr error) {
  var err error 
  base.Debug("[objAPIServer.UpdateObject] parent Dir, Object hash: ", parentDir, objectHash) 


  cksum_hash := utils.NewHash()  
  r := io.TeeReader(fileReader, cksum_hash)  

  objPath = objectHash

  wc, err := s.storageClient.Writer(objPath) 

  if err != nil {
    retErr = err
    return 
  } 

  defer func() {
    if wc != nil {
      err := wc.Close()
      if err != nil && retErr == nil{
        base.Log("[objAPIServer.UpdateObject] Error closing writer: ", parentDir, objectHash, err)
        retErr = err
      }
    }  
  }()

  
  if inchunks {

    written, err = io.CopyN(wc, r, splitSize)
    _, _ = written, err
  
  } else {

    // create and use own buffer queue to limit memory usage
    buffer := utils.GetBuffer()
    defer utils.PutBuffer(buffer)

    written, err = io.CopyBuffer(wc, r, buffer)
    
    if err != nil {
      retErr = err
      return
    }  

    base.Debug("[objApiServer.UpdateObject] Object Path and Write Size - ", objPath, written)
    
    new_size, err := s.storageClient.Size(objectHash)
    if err != nil {
      base.Debug("[objApiServer.UpdateObject] Failed to retrieve Object size from storage bucket: ", objPath)
    }

    base.Debug("[objApiServer.UpdateObject] New size of object: ", new_size)
    
    _, _ = written, err
  }
  
  if err != nil {
    if err != io.EOF {
      // delete half created object
      s.storageClient.Delete(objectHash)
      retErr = err
      return
    }
    if err == io.EOF && written == 0 {
      base.Debug("[objApiServer.UpdateObject] Reached EOF. No Writes. obj Path:", objPath)
      return
    }
  }

  // calculate checksum 
  writesum = utils.HexToString(cksum_hash.Sum(nil))
  return  
}

func (s *objAPIServer) DeleteObject(path string) error {
  return s.storageClient.Delete(path)
}

func (s *objAPIServer) GetObjectSize(name string) int64 {
  obj_size, err := s.storageClient.Size(name)
  if err != nil {
    return 0
  }

  return obj_size
}

//TODO: add cancel with context 
//
func (s *objAPIServer) GetObject(name string, offset int64, size int64) ([]byte, int, error) {
	var err error
  var rc io.ReadCloser 

	rc, err = s.storageClient.Reader(name, offset , size)

	if (err != nil) {
    //TODO: temporary hack for now to detect end of file
    if strings.Contains(err.Error(), "InvalidRange") {
      fmt.Println("reached end of file")
    }
    fmt.Println("error", err)
    return nil, 0, io.EOF
	}

	data, err := ioutil.ReadAll(rc) 

	if (err != nil) {
		return nil, 0, err
	}


	return data, len(data), nil

}

 

//  Read object into a writer 
func (s *objAPIServer) ReadObject(objPath string, offset int64, size int64, objWriter io.Writer ) (int64, error) {
  var err error
  var bytesWritten int64
  
  // TODO: read & write in chunk
  fmt.Println("This is reader call", offset, size)
  
  r, err  := s.Reader(objPath, offset , size)

  //TODO: handle io.EOF on caller

  if err!= nil {
    fmt.Println("error", err)
    return 0, err
  }

  defer r.Close()

  buffer := utils.GetBuffer()
  defer utils.PutBuffer(buffer)
  bytesWritten, err = io.CopyBuffer(objWriter, r, buffer)

  if err != nil {
    if err != io.EOF {
      return 0, err
    }
  }

  return bytesWritten, err
}

func (s *objAPIServer) CheckObject(name string) bool {
  fmt.Println("name", name)
  found := s.storageClient.Exists(name)

  return found 
}





package storage

import (
  "fmt"
  "io"
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
  PutObject(parentDir string, fileReader io.Reader, inchunks bool) (string, string, int64, error)
  GetObject(name string) ([]byte, int, error) 
  CheckObject(name string) bool
  DeleteObject(path string) error
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

func (s *objAPIServer) objDir(path string) string {
  return filepath.Join(s.dir, path)
}

func (s *objAPIServer) objPath(path string, file_name string) string {
  return filepath.Join(s.objDir(path), file_name)
}

// TODO: add error channel and submit a go func to write object
// TODO: add cancel with context
//
func (s *objAPIServer) PutObject(parentDir string, fileReader io.Reader, inchunks bool) (string, string, int64, error) {
	var err error 

  cksum_hash := utils.NewHash()

  obj_hash := utils.NewUUID()
  obj_path := s.objPath(parentDir, obj_hash)
  
  r := io.TeeReader(fileReader, cksum_hash) 
  
	wc, err := s.storageClient.Writer(obj_path) 
  
  if err != nil {
    return obj_path,"", 0, err
  }

  defer wc.Close()
  
  var n int64
  if inchunks {
    n, err = io.CopyN(wc, r, splitSize)
    _, _ = n, err
  } else {
    // create and use own buffer queue to limit memory usage
    buffer := utils.GetBuffer()
    defer utils.PutBuffer(buffer)
    n, err = io.CopyBuffer(wc, r, buffer)
    _, _ = n, err
  }
  
  if err != nil {
    if err != io.EOF {
      // delete half created object
      s.storageClient.Delete(obj_path)
    }
    return obj_path,"", 0, err
  }

  // calculate checksum 
  obj_checksum := utils.HexToString(cksum_hash.Sum(nil))
  return obj_path, obj_checksum, n, err
}


func (s *objAPIServer) DeleteObject(path string) error {
  return s.storageClient.Delete(path)
}

//TODO: add cancel with context 
//
func (s *objAPIServer) GetObject(name string) ([]byte, int, error) {
	var err error

	rc, err  := s.storageClient.Reader(name, 0 , 0)

	if (err != nil) {
		return nil, 0, err
	}

	data, err := ioutil.ReadAll(rc)

	if (err != nil) {
		return nil, 0, err
	}


	return data, len(data), nil

}

func (s *objAPIServer) CheckObject(name string) bool {
  fmt.Println("name", name)
  found := s.storageClient.Exists(name)

  return found 
}





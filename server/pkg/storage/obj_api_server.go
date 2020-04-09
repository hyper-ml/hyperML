package storage

import (
  "fmt"
  "io"
  "hash"
  "strings"
  "io/ioutil"  

  "path/filepath" 

  "github.com/hyper-ml/hyperml/server/pkg/utils"
  "github.com/hyper-ml/hyperml/server/pkg/base"
  "github.com/hyper-ml/hyperml/server/pkg/config"
)

type ObjectType string
const (
  Object ObjectType = "objects"
  Log ObjectType = "logs"
)

var (
  objSplitSize = int64(16 * 1024 * 1024)
)


type ObjectAPIServer interface {
  BasePath() string
  GetObjectPath(name string) string

  Reader(path string, offset int64, size int64) (io.ReadCloser,error)
  Writer(path string) (io.WriteCloser, error) 
  ReadSeeker(path string, offset int64, size int64) (io.ReadSeeker, error)
  
  NewObjectHash() string
  CreateObject(fileReader io.Reader, inchunks bool) (string, string, int64, error)
  SaveObject(name string, subdir string, fileReader io.Reader, inchunks bool) (path string, chksm string, written int64, fnerr error) 
  GetObject(name string, offset int64, size int64) ([]byte, int, error) 

  ObjectURL(op, name string) (string, error)
  ObjectSize(name string) (int64)
  GetSizeByPath(path string) int64

  CheckObject(path string) bool
  DeleteObject(path string) error

  MergeObjects(dest string, src []string) error
}
 
type objAPIServer struct {
  basePath string
	objectType ObjectType
	handler Client
}

func NewObjectAPI(c *config.ObjStorageConfig) (*objAPIServer, error) {
  return NewTypedAPIServer(c, Object)
}

func NewLogObjectAPI(c *config.ObjStorageConfig) (*objAPIServer, error) {
  return NewTypedAPIServer(c, Log)
}

func NewTypedAPIServer(c *config.ObjStorageConfig, kind ObjectType) (*objAPIServer, error){
  var base_path string
  
  client, err := NewClient(c)
  if err != nil {
    return nil, fmt.Errorf("failed to create a storage client: %v", err)
  }

  if (kind == Log) {
      base_path = filepath.Join(c.BaseDir, string(Log))
  } else {
      base_path = filepath.Join(c.BaseDir, string(Object))
  }

  return &objAPIServer{
    basePath: base_path,
    objectType: kind,
    handler: client,
  }, nil

}

func (s *objAPIServer) BasePath() string {
  return s.basePath
} 

func (s *objAPIServer) GetObjectPath(name string) string { 
  return filepath.Join(s.basePath, name)
}
 
func (s *objAPIServer) NewObjectHash() string {
  return utils.NewUUID()
}

func (s *objAPIServer) newCheckSumHash() hash.Hash {
  return utils.NewHash()
}

func (s *objAPIServer) CheckObject(path string) bool {
  found := s.handler.Exists(path)
  return found 
}

func (s *objAPIServer) GetObject(path string, offset int64, size int64) (data []byte, new_size int, fnerr error) {
 
  var r io.ReadCloser 

  r, fnerr = s.handler.Reader(path, offset , size)
  if (fnerr != nil) {
    if strings.Contains(fnerr.Error(), "InvalidRange") {
      fnerr = io.EOF
      return 
    }
    return  
  }

  data, fnerr = ioutil.ReadAll(r) 
  if (fnerr != nil) {
    return 
  }

  return data, len(data), nil
}

func (s *objAPIServer) CreateObject(fileReader io.Reader, inchunks bool) (string, string, int64, error) {
  new_hash := s.NewObjectHash()
  return s.SaveObject(new_hash, "", fileReader, inchunks)
}

// save object 
func (s *objAPIServer) SaveObject(name string, subdir string, fileReader io.Reader, inchunks bool) (path string, chksm string, written int64, fnerr error) {
  
  // new checksum hash
  cksum_hash := s.newCheckSumHash()

  // derive direct object path
  path = s.GetObjectPath(subdir)
  path = filepath.Join(path, name)

  // reader and writer
  r := io.TeeReader(fileReader, cksum_hash)   
  wc, fnerr := s.handler.Writer(path) 

  if fnerr != nil { 
    return  
  }

  if inchunks {
    fnerr = fmt.Errorf("[objAPIServer.SaveObject] feature_not_implemented: inchunks", path)  
    return
  } else {

    // get access to a buffer 
    buffer := utils.GetBuffer()
    defer utils.PutBuffer(buffer)
    
    // write buffer 
    written, fnerr = io.CopyBuffer(wc, r, buffer)
    
  }

  if fnerr != nil {
    if fnerr != io.EOF {
 
      // delete partially created object
      s.handler.Delete(path)
      base.Error("[objAPIServer.CreateObject] Failed to write object %s: %v", path, fnerr)
      return
 
    }

    if fnerr == io.EOF && written == 0 {
      return
    }
  }

  if err := wc.Close(); err != nil {
    fnerr = err 
    return 
  }  

  // calculate checksum 
  chksm = utils.HexToString(cksum_hash.Sum(nil))
  return 
}


func (s *objAPIServer) Writer(objpath string) (io.WriteCloser, error) {
  return s.handler.Writer(objpath)
}

func (s *objAPIServer) Reader(objpath string, offset int64, size int64) (io.ReadCloser,error) {
  reader, err := s.handler.Reader(objpath, offset , size)
  
  if (err != nil) {
    //TODO: temporary hack for now to detect end of file
    if strings.Contains(err.Error(), "InvalidRange") {
       return nil, io.EOF
    } 
    return nil, err
  }

  return reader, err
}

func  (s *objAPIServer) ReadSeeker(objpath string, offset int64, size int64) (io.ReadSeeker, error) {

  var object_size int64
  reader, err := s.Reader(objpath, offset, size)
  
  if err != nil {
    return nil, err
  }

  if size != 0 {
    object_size = size 
  } else {
    object_size = s.GetSizeByPath(objpath) //ObjectSize
  }

  return &getObjReadSeeker{
    reader: reader,
    obj_path: objpath,
    offset: offset,
    size: object_size,
    s: *s,
  }, nil
}

func (s *objAPIServer) DeleteObject(path string) error {
  return s.handler.Delete(path)
}

func (s *objAPIServer) ObjectSize(name string) int64 {
  return s.GetSizeByPath(s.GetObjectPath(name))
}

func (s *objAPIServer) GetSizeByPath(path string) int64 {
  obj_size, err := s.handler.Size(path)
  if err != nil {
    return 0
  }

  return obj_size
}  

func (s *objAPIServer) ObjectURL(op, name string) (string, error){
  name = s.GetObjectPath(name)
  return s.handler.SignedURL(op, name)
}

func (s *objAPIServer) MergeObjects(dest string, src []string) error {
  return s.handler.Merge(s.BasePath(), dest, src)
}
   





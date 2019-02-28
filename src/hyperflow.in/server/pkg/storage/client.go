package storage


import (
  "io"
  "fmt"
  "golang.org/x/net/context"

  config_pkg "hyperflow.in/server/pkg/config"
)


type Client interface {
  // Writer returns a writer which writes to an object.
  Writer(name string) (io.WriteCloser, error)

  // Reader returns a reader which reads from an object.
  Reader(name string, offset int64, size int64) (io.ReadCloser, error)

  // Delete deletes an object.
  Delete(name string) error

  // Size of the object
  Size(name string) (int64, error)

  // Check if object exists
  Exists(name string) bool

  // IsNotExist returns true if err is a non existence error
  IsNotExist(err error) bool

  // return a signed url for given http method
  SignedURL(op, objname string) (string, error)

  // merge objects 
  Merge(parentDir, dest string, src []string) error
}

func NewClient(objConfig *config_pkg.ObjStorageConfig) (Client, error) {
  switch {
  case objConfig.StorageTarget == config_pkg.GCS:
    return NewGcsClient(objConfig)
  case objConfig.StorageTarget == config_pkg.S3:
    return NewS3Client(objConfig) 

  default:
    fmt.Errorf("[storage.NewClient] Incorrect Storage config")
  }
  return nil, nil
}

func NewGcsClient(c *config_pkg.ObjStorageConfig) (Client, error) {
  return newGcsClient(context.Background(), c)
}

func NewS3Client(c *config_pkg.ObjStorageConfig) (Client, error) {
  return newS3Client(c)
}







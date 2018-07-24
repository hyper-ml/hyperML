package storage


import (
  "io"
  "golang.org/x/net/context"

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
}


// NewGoogleClient creates a google client with the given bucket name.
func NewGoogleClient(ctx context.Context, bucket string) (Client, error) {
  return newGoogleClient(ctx, bucket)
}


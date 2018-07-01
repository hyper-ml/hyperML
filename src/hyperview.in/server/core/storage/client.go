package storage


import (
  "io"
  "golang.org/x/net/context"

)


type Client interface {
  // Writer returns a writer which writes to an object.
  // It should error if the object already exists or we don't have sufficient
  // permissions to write it.
  Writer(name string) (io.WriteCloser, error)

  // Reader returns a reader which reads from an object.
  // If `size == 0`, the reader should read from the offset till the end of the object.
  // It should error if the object doesn't exist or we don't have sufficient
  // permission to read i
  Reader(name string, offset uint64, size uint64) (io.ReadCloser, error)

  // Delete deletes an object.
  // It should error if the object doesn't exist or we don't have sufficient
  // permission to delete it.
  Delete(name string) error

  // Exsits checks if a given object already exists
  Exists(name string) bool

  // IsNotExist returns true if err is a non existence error
  IsNotExist(err error) bool
}


// NewGoogleClient creates a google client with the given bucket name.
func NewGoogleClient(ctx context.Context, bucket string) (Client, error) {
  return newGoogleClient(ctx, bucket)
}


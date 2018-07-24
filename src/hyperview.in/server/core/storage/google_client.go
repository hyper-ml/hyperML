package storage

import (
	"io" 

	gstore "cloud.google.com/go/storage"
	"golang.org/x/net/context"

  "hyperview.in/server/base"
)


type googleClient struct {
	ctx    context.Context
	bucket *gstore.BucketHandle
}


func newGoogleClient(ctx context.Context, bucket string) (*googleClient, error) {
	client, err := gstore.NewClient(ctx)
	if err != nil {
		return nil, err
	}
  base.Debug("[newGoogleClient] Creating bucket pointer", bucket)
	return &googleClient{ctx, client.Bucket(bucket)}, nil
}
 

// TODO:  add context, retry exponential back-off 
func (c *googleClient) Writer(name string) (io.WriteCloser, error) {
	return c.bucket.Object(name).NewWriter(c.ctx), nil
}

func (c *googleClient) Reader(name string, offset int64, size int64) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error
	if size == 0 {
		// a negative length will cause the object to be read till the end
		reader, err = c.bucket.Object(name).NewRangeReader(c.ctx, offset, -1)
	} else {
		reader, err = c.bucket.Object(name).NewRangeReader(c.ctx, offset, size)
	}
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (c *googleClient) Delete(name string) error {
	return c.bucket.Object(name).Delete(c.ctx)
}

//object exists
func (c *googleClient) Exists(name string) bool {
	_, err := c.bucket.Object(name).Attrs(c.ctx)
  
  if err != nil {
    base.Debug("[newGoogleClient.Exists] Failure while checking if object exists", err)
  }

  return err == nil
}

func (c *googleClient) Size(name string) (int64, error) {
	objAttrs, err := c.bucket.Object(name).Attrs(c.ctx)

	if err != nil {
    base.Debug("[newGoogleClient.Size] Failure while calling Object.Attrs()", err)
    return int64(0), err
  }
  //fmt.Println("object attributes", objAttrs)
  return objAttrs.Size, nil
}

//object doesnt exist ?
func (c *googleClient) IsNotExist(err error) (result bool) {
	return err == gstore.ErrObjectNotExist
}


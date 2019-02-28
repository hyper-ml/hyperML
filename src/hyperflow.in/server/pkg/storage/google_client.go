package storage

import (
	"io" 
  "fmt"
  "time"
  "io/ioutil"
  filepath_pkg "path/filepath"
	gstorage "cloud.google.com/go/storage"
	"golang.org/x/net/context"
  "golang.org/x/oauth2/jwt"

  "golang.org/x/oauth2/google"
  "google.golang.org/api/option"

  config_pkg "hyperflow.in/server/pkg/config"
  "hyperflow.in/server/pkg/base"
)


type gcs struct {
  ctx context.Context
  bucket_name string
	bucket *gstorage.BucketHandle
  accessID string
  pkey []byte
}


func newGcsClient(ctx context.Context, objConfig *config_pkg.ObjStorageConfig) (Client, error) {

  var creds []byte
  var client *gstorage.Client
  var err error 

  cnf := objConfig.Gcs
  base.Println("GCS Storage Cloud Bucket: ", cnf.Bucket)

  switch {
  case cnf.Creds != nil:
    creds = cnf.Creds
  case cnf.CredsPath != "":
    creds, err = readCredPath(cnf.CredsPath)
    if err != nil {
      return nil, fmt.Errorf("failed to read cred path from %s, err: %v", cnf.CredsPath, err)
    }
  default: 
    return nil, fmt.Errorf("missing google storage (GCS) credentials")
  }

  creds_obj, err := google.CredentialsFromJSON(ctx, creds, gstorage.ScopeReadWrite)
  if err != nil {
    return nil, fmt.Errorf("failed to read google creds from JSON, err: %v", err)
  }

  client, err = gstorage.NewClient(ctx, option.WithCredentials(creds_obj))

	if err != nil {
		return nil, err
	}

  // read json
  jwt_config, _ := jwtConfig(creds) 
  
  return &gcs { 
    ctx: ctx,
    bucket_name: cnf.Bucket,
    bucket: client.Bucket(cnf.Bucket),
    accessID: jwt_config.Email,
    pkey: jwt_config.PrivateKey,
  }, nil
}
 
func (c *gcs) SignedURL(op, objname string) (string, error) {
  return gstorage.SignedURL(c.bucket_name, objname, 
    &gstorage.SignedURLOptions{
      GoogleAccessID: c.accessID,
      PrivateKey:     c.pkey,
      Method:         op,
      Expires:        time.Now().Add(1 * time.Hour),
  })
} 

// TODO:  add context, retry exponential back-off 
func (c *gcs) Writer(name string) (io.WriteCloser, error) {
	return c.bucket.Object(name).NewWriter(context.Background()), nil
}

func (c *gcs) Reader(name string, offset int64, size int64) (io.ReadCloser, error) {
	var reader io.ReadCloser
  ctx := context.Background()

	var err error
	if size == 0 {
		// a negative length will cause the object to be read till the end
		reader, err = c.bucket.Object(name).NewRangeReader(ctx, offset, -1)
	} else {
		reader, err = c.bucket.Object(name).NewRangeReader(ctx, offset, size)
	}
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (c *gcs) Delete(name string) error {
	return c.bucket.Object(name).Delete(context.Background())
}

//object exists
func (c *gcs) Exists(name string) bool {
	_, err := c.bucket.Object(name).Attrs(context.Background())
  
  if err != nil {
    base.Debug("[newGoogleClient.Exists] Failure while checking if object exists", err)
  }

  return err == nil
}

func (c *gcs) Size(name string) (int64, error) {
	objAttrs, err := c.bucket.Object(name).Attrs(context.Background())

	if err != nil {
    base.Error("object %s size find error: %v", name, err)
    return int64(0), fmt.Errorf("failed to find object %s, err: %v", name, err)
  }
  
  return objAttrs.Size, nil
}

//object doesnt exist ?
func (c *gcs) IsNotExist(err error) (result bool) {
	return err == gstorage.ErrObjectNotExist
}

func (c *gcs) Merge(parentDir string, dest string, src []string) error{
  
  dest_path := filepath_pkg.Join(parentDir, dest)
  dest_handle := c.bucket.Object(dest_path)
  
  var src_handles []*gstorage.ObjectHandle
  src_handles = append(src_handles, dest_handle)

  for _, hash := range src {
    hash_path := filepath_pkg.Join(parentDir, hash)
    src_handles = append(src_handles, c.bucket.Object(hash_path))
  }
  
  cmpsr := dest_handle.ComposerFrom(src_handles...)
  if _, err := cmpsr.Run(c.ctx); err != nil {
    return fmt.Errorf("failed to merge file paths, err: ", err)
  }

  return nil
}

// utility funcs 
//
func readCredPath(path string) ([]byte, error){
  return ioutil.ReadFile(path)
}

func jwtConfig(creds []byte) (*jwt.Config, error) {

  cf, err := google.JWTConfigFromJSON(creds, gstorage.ScopeReadWrite)
  if err != nil {
    return nil, fmt.Errorf("google.JWTConfigFromJSON: %v", err)
  }

  return cf, nil
}
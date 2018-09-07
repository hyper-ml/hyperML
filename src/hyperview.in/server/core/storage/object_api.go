package storage

import (
  "fmt"

  "hyperview.in/server/base"
)

// Valid object storage
const ( 
  GoogleStorage    = "GCS" 
)

type StorageServer interface {
	ObjectAPIServer
}

func NewObjectAPI(dir string, cacheBytes int64, storage_backend string) (ObjectAPIServer, error) {
  switch storage_backend {
    case GoogleStorage:
      o, err := newGoogleStorageAPIServer(dir, cacheBytes)

      if err != nil {
        base.Log("[storage.NewObjectAPI] Failed to create storage API: ", err)
        return nil, err
      }

      return o, nil

  }
  return nil, fmt.Errorf("Unknown Storage Location")

}

func NewBucketAPI(bucket string, dir string, storage_backend string) (ObjectAPIServer, error) {
  switch storage_backend {
    case GoogleStorage:
      o, err := newGoogleBucketAPIServer(bucket, dir, 0)

      if err != nil {
        base.Log("[storage.NewBucketAPI] Failed to create storage API: ", err)
        return nil, err
      }

      return o, nil

  }
  return nil, fmt.Errorf("Unknown Storage Location")


}

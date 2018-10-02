package storage

import (
  "fmt"

  "hyperview.in/server/base"
)

// Valid object storage
const ( 
  GoogleStorage    = "GCS" 
  AwsStorage       = "S3"
)

type StorageServer interface {
	ObjectAPIServer
}

func NewObjectAPI(baseDir string, storageOption string) (ObjectAPIServer, error) {
  switch storageOption {
    case GoogleStorage:
      o, err := newGoogleStorageAPIServer(baseDir)

      if err != nil {
        base.Log("[storage.NewObjectAPI] Failed to create storage API: ", err)
        return nil, err
      }

      return o, nil

  }
  return nil, fmt.Errorf("Unknown Storage Location")

}

func NewBucketAPI(bucket string, baseDir string, storageOption string) (ObjectAPIServer, error) {
  switch storageOption {
    case GoogleStorage:
      o, err := newGoogleBucketAPIServer(bucket, baseDir)

      if err != nil {
        base.Log("[storage.NewBucketAPI] Failed to create storage API: ", err)
        return nil, err
      }

      return o, nil

  }
  return nil, fmt.Errorf("Unknown Storage Location")


}

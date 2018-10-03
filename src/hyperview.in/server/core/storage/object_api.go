package storage

import (
  "fmt"

  "hyperview.in/server/base"
  "hyperview.in/server/config"
)



type StorageServer interface {
	ObjectAPIServer
}

func NewObjectAPI(storageOption string, baseDir string, s3Config *config.S3Config, gConfig *config.GcsConfig) (ObjectAPIServer, error) {
  switch storageOption {
    case config.GStorage:
      o, err := newGoogleStorageAPIServer(baseDir, gConfig)

      if err != nil {
        base.Log("[storage.NewObjectAPI] Failed to create storage API: ", err)
        return nil, err
      }

      return o, nil

  }
  return nil, fmt.Errorf("Unknown Storage Location")

} 

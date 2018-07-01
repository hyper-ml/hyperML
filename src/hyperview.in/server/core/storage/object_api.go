package storage

import ("fmt")

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
        return nil, err
      }

      return o, nil

  }
  return nil, fmt.Errorf("Unknown Storage Location")

}


func NewStorageServer(dir string, cacheBytes int64, storage_backend string) (StorageServer, error) {
	switch storage_backend {
    case GoogleStorage:
      s, err := newGoogleStorageAPIServer(dir, cacheBytes)
      if err != nil {
        return nil, err
      }

      return s, nil

  }
  return nil, fmt.Errorf("Unknown Storage Location")

}
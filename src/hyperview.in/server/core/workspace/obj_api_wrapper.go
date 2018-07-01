package workspace

import (
  "io"
  "hyperview.in/server/core/storage"
)

type ObjApiWrapper struct {
	api storage.ObjectAPIServer
}

const (
	OBJECT_DIR = "objects"
)

func (o *ObjApiWrapper) PutObject(commit *Commit, filePath string, reader io.Reader, inChunks bool) (*FileInfo, error) {
  obj_path, check_sum, bytes, err := o.api.PutObject(OBJECT_DIR, reader, inChunks)
  if err != nil {
    return nil, err
  }

  return NewFileInfo(commit, filePath, obj_path, bytes, check_sum), nil
}





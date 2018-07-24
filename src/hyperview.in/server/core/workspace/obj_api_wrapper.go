package workspace

import (
  "io" 
  "strings"

  "hyperview.in/server/base"
  "hyperview.in/server/core/storage"
)

type ObjWrapper struct {
	api storage.ObjectAPIServer
}

const (
	OBJECT_DIR = "objects"
)

// write  to storage server 
// needs objecthash, reader and let the magic being
// TODO: needs to do something about concurrent writes to same file 
// check sum is useless unless sent by client too 
// for now just compare the size 
func (o *ObjWrapper) AppendObject(objecthash string, reader io.Reader) (string, string, int64, error) {
  
  

  obj_path, check_sum, bytes, err := o.api.UpdateObject(OBJECT_DIR, objecthash, reader, false)

  if err != nil {
    base.Log("[ObjWrapper.AppendObject] Update Object Failed with err: ", err)
    return "", "", 0, err
  }

  return obj_path, check_sum, bytes, nil
}

func (o *ObjWrapper) CreateObject(reader io.Reader) (string, string, int64, error) {
  new_hash, check_sum, bytes, err := o.api.CreateObject(OBJECT_DIR, reader, false)
  
  base.Debug("New Storage object created:", new_hash, bytes)

  if err != nil {
    return "", "", 0, err
  }

  return new_hash, check_sum, bytes, nil  
}



func (o *ObjWrapper) CreateEmpty() (string, int64, error) {
  obj_path, _, bytes, err := o.api.CreateObject(OBJECT_DIR, strings.NewReader(""), false)
  if err != nil {
    return "", 0, err
  }

  return obj_path, bytes, nil  
}





package storage

import (
  "io" 
  "fmt"
)

type reader = io.Reader

//used for serving http content 
type getObjReadSeeker struct {
  reader 
  obj_path string
  offset int64
  size int64  
  s objAPIServer
}

func (ors *getObjReadSeeker) Seek(offset int64, whence int) (int64, error) {
  offsetReader := func(offset int64) (io.Reader, error) {
    return ors.s.Reader(ors.obj_path, offset, 0)
  }
  
  switch whence {
    case io.SeekStart:
      reader, err := offsetReader(offset)
      if err != nil {
        return ors.offset, err
      }
      ors.offset = offset
      ors.reader = reader

    case io.SeekCurrent:
      reader, err := offsetReader(ors.offset + offset)
      if err != nil {
        return ors.offset, err
      }
      ors.offset += offset
      ors.reader = reader
    
    case io.SeekEnd: 
      reader, err := offsetReader(ors.size - offset)
      if err != nil {
        if err != io.EOF { 
        return ors.offset, err
        }
      }
      ors.offset = ors.size - offset
      ors.reader = reader
  }  
  fmt.Println("in seek", ors.offset, offset, whence)
  return ors.offset, nil
} 
package storage


import (
  "os"
  "io" 
  . "bytes"
  "testing"
  "fmt"
  "strings"
)

const (
  TEST_DIR = "test_dir"
  STORAGE_BACKEND = "GCS"
  TEST_CHUNKSIZE = int64(16 * 1024 * 1024) // 16 MB
)

func setEnv(){
  os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/apple/MyProjects/creds/hw-storage-75d060e8419a.json")
  os.Setenv("GOOGLE_STORAGE_BUCKET", "hyperview001")
}

func TestMain(m *testing.M) {
  setEnv()
  exit := m.Run()
  os.Exit(exit)
}


func Test_ObjectWrite(t *testing.T) {

  objectAPI, err:= NewObjectAPI(TEST_DIR, 0, STORAGE_BACKEND)

  file_path, csum, bwrt, err := objectAPI.CreateObject("objects", strings.NewReader("Hello"), false)
  
  fmt.Println("file_path, checksum, bytes: ", file_path, csum, bwrt)

  if err != nil {
    fmt.Println("error put object: \n", err)
  } 

  fmt.Println("Sent object. Bytes written:", file_path, bwrt)

  b := objectAPI.CheckObject(file_path)

  if (!b) {
  	t.Fatalf("Unexpected Error in creation of object in objectCreate()")
  }
  fmt.Println("Deleting object:", file_path)
  err = objectAPI.DeleteObject(file_path)
  if err != nil {
    fmt.Println("Failed to delete object from server:", err)
    t.Fatalf("Failed to delete object from server")
  }
}
/*
func Test_ReadObject(t *testing.T) {
  var data Buffer

  api, err:= NewObjectAPI(TEST_DIR, 0, STORAGE_BACKEND)

  file_path, _, _, err := api.PutObject("objects", strings.NewReader("Hello"), false)
  if err != nil {
    fmt.Println("error in put object: \n", err)
    t.Fatalf("Failed to write file on server for reading")
  }

  _, err = api.ReadObject(file_path, &data)
  if err != nil {
    fmt.Println("Read object failed", err)
  }

  fmt.Println("size of data:", data.Len())
  fmt.Println("data:", &data)


}*/

func putTestFile(api ObjectAPIServer) (string, error) {
  file_path, _, _, err := api.CreateObject("objects", strings.NewReader("Hello World. Testing text files. Short files."), false)
  if err != nil {
    fmt.Println("error in put object: \n", err)
    return "", err
  }
  return file_path, err
} 


func Test_ReadObjectInChunks(t *testing.T) {
  var data Buffer
  var file_path string

  api, err:= NewObjectAPI(TEST_DIR, 0, STORAGE_BACKEND)

  var size uint64
  var offset uint64
  var n int64 

  file_path, err = putTestFile(api)
  
  offset = 0
  size = 1
  var block_size uint64 = 1

  for {

    n, err = api.ReadObject(file_path, int64(offset), int64(size), &data) 
    if err == io.EOF {
      break
    } 

    offset = offset +  block_size
  }
  var byt byte
  str, err := data.ReadString(byt)

  fmt.Println("data: ", str)
  fmt.Println("size read:", n)

  if err!= nil && err != io.EOF {
    t.Fatalf("Error occured during seek %d %d",offset, size)
  }
}

 




package storage


import (
  "os"
  "testing"
  "fmt"
  "strings"
)

const (
  TEST_DIR = "test_dir"
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


func TestIntegration_ObjectOps(t *testing.T) {

  objectAPI, err:= NewObjectAPI(TEST_DIR, 0, "GCS")

  file_path, csum, bwrt, err := objectAPI.PutObject("objects", strings.NewReader("Hello"), false)
  
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

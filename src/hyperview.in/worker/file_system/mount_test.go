package file_system

import (
  "os"
  "testing"
  log "github.com/sirupsen/logrus"
  "fmt"

)

func initLog() {
  log.SetFormatter(&log.JSONFormatter{})
  log.SetLevel(log.WarnLevel)
  log.SetOutput(os.Stdout)

}

func Test_Mount(t *testing.T) {
  err:= mount("test_repo", "/", "/Users/apple/MyProjects/stash/hw")
  if err != nil {
    t.Fatalf("Failed to mount test repo: %s", err)
  }
}

/// not working TODO 
func Test_Unmount(t *testing.T) {
  Unmount( "/Users/apple/MyProjects/stash/hw")
}

func Test_CreateFile(t *testing.T) {
  f, err := os.Create("/Users/apple/MyProjects/stash/hw/p.py")
  if err != nil {
    fmt.Println("[Test_create] Failed to create file: ", err)
  }
  defer f.Close()
}
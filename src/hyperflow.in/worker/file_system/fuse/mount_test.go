package fuse

import (
  "os"
  "time"
  "testing"
  "fmt"


  log "github.com/sirupsen/logrus"

  "hyperflow.in/worker/api_client"
)

const (
  TEST_WORKSPACE_DIR = "/Users/apple/MyProjects/stash/hw"
  TEST_REPO_NAME = "test_repo0123456789"
  TEST_LOCAL_MOUNT_PATH ="/"
)

func initLog() {
  log.SetFormatter(&log.JSONFormatter{})
  log.SetLevel(log.WarnLevel)
  log.SetOutput(os.Stdout)

}

func Test_Mount(t *testing.T) {
  
  wc, err := api_client.NewWorkerClient("")

  if err != nil {
    fmt.Println("[Test_Mount] Failed to launch Worker Client: ", err)
    t.Fail()
  }

  mnt, err := mount(TEST_REPO_NAME, TEST_LOCAL_MOUNT_PATH, TEST_WORKSPACE_DIR, wc)
  if err != nil {
    t.Fatalf("Failed to mount test repo: %s", err)
  }
  time.Sleep(200 * time.Second)
  mnt.Close()
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
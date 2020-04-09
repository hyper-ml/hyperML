package test_util

import(
  "os"
  "fmt"
  "testing"
)


func createLargeFile(path string) error {
  f, err := os.Create(path)
  if err != nil {
    return fmt.Errorf("failed to create large file, err: %v", err)
  }
  
  if err := f.Truncate(1e7); err != nil {
    return err
  }

  return nil
}
 

func Test_PushDir(t *testing.T) {
  tc, err := NewTestConfig()
  defer func() {
    _ = tc.Destroy()
  }()

  if err != nil {
    t.Fatalf("failed to generate test config ,err: %v", err)
  }  

  if err = tc.SetRepoFs(); err != nil {
    t.Fatalf("failed to set repo fs, err: %v", err)
  } 

  sent, err := tc.RepoFs.PushDir("") 
  if err != nil {
    t.Fatalf("failed to push object, err: %v", err)
  }

  if sent != int64(len(TestCodeScript)) {
    t.Fatalf("failed to push repo %d %d", sent, len(TestCodeScript))
  }

}

func Test_PushDirLarge(t *testing.T) {
  tc, err := NewTestConfig()
  defer func() {
    _ = tc.Destroy()
  }()

  if err != nil {
    t.Fatalf("failed to generate test config ,err: %v", err)
  }  

  if err := tc.SetRepoFs(); err != nil {
    t.Fatalf("failed to set repo fs, err: %v", err)
  }

  if err := createLargeFile(tc.RepoFs.GetLocalFilePath("large.py")); err != nil {
    t.Fatalf("err: %v", err)
  }

  sent, err := tc.RepoFs.PushDir("") 
  if err != nil {
    t.Fatalf("failed to push object, err: %v", err)
  }

  if sent == 0 {
    t.Fatalf("failed to push repo %d", sent)
  }
}
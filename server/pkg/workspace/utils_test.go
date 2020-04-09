package workspace_test

import(
  "fmt"
  "testing"
  ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
)

///parent/child1/child2/0.mp3", "/parent/1.mp3"
func Test_ListToDir(t *testing.T){
  m, err := ws.ListToD([]string{"/parent1/child1/232", "/parent1/child1/child2/0.mp3", "train.py", "/parent/1.mp3"})
  fmt.Println("m: ", m["/"], err)
}
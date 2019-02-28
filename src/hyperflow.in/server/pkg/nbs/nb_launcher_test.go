package nbs_test


import(
  "fmt"
  "testing"
  "hyperflow.in/server/nbs"
)


func Test_newNote(t *testing.T) {
  l := nbs.NewNbLauncher()
  err:= l.NewNotebook()
  if err != nil {
    fmt.Println("Error starting notebook: ",err)
    t.Fail()
  }
}
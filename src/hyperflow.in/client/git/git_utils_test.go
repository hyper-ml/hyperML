package git


import(
  "testing"
)


func Test_InitGitRepo(t *testing.T){
  r, e := Init("/Users/apple/MyProjects/stash/git1")
  fmt.Println("e: ", e)
  fmt.Println("r: ", r)
}
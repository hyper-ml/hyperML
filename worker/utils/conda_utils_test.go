package utils_test


import(
  "testing"
  "github.com/hyper-ml/hyperml/worker/utils"
)


func Test_CreateCondaEnvBySpec(t *testing.T) {
  _, err := utils.GetOrCreateCondaEnvWithSpec("/Users/apple/MyProjects/stash/work1/conda.yml")
  if err != nil {
    t.Fatalf("Failed to create conda environment: %s", err)
  }
}



func Test_CreateCondaDefaultEnv(t *testing.T) {
  _, err := utils.GetOrCreateDefaultCondaEnv("test_p1")
  if err != nil {
    t.Fatalf("Failed to create conda environment: %s", err)
  }
}





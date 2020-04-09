package deploy

import(
  "fmt"
  "io/ioutil"
  "testing"
)


func Test_BaseConfig(t *testing.T) {
  manifest, err := BaseConfig(nil, nil)
  if err != nil {
    t.Fatalf("failed creating base config, err: %v", err)
  }

  b:= manifest.ReadCacheBytes()
  
  err = manifest.FlushToFile("")
  if err != nil {
    fmt.Println("err: ", err)
  }

  fmt.Println(string(b))
}

func Test_GoogleConfig(t *testing.T) {
  
  g_creds_path := "/Users/apple/MyProjects/hf_src_build/server/gcloud_config.json"
  creds, _ := ioutil.ReadFile(g_creds_path)

  manifest, err := GoogleConfig(nil, nil, string(creds), "hyperflow001", None, "", "", "/var/tmp", "50Mi")
  if err != nil {
    t.Fatalf("failed creating base config, err: %v", err)
  }

  b:= manifest.ReadCacheBytes()
  
  err = manifest.FlushToFile("")
  if err != nil {
    fmt.Println("err: ", err)
  }

  fmt.Println(string(b))
}

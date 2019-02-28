package flow_test

import(
  "io"
  "fmt"
  "hyperflow.in/server/pkg/flow"
  "testing"
)


func Test_LogStream(t *testing.T) {
  flow_id := "9d9f19754e714d2fbad352c05a33ecba"
  pk := flow.NewDefaultPodKeeper(nil, nil)
  ro, err := pk.LogStream(flow_id)
  if err != nil {
    t.Fatalf("error: %s", err)
  }
  
  var n int 
  p := make([]byte, 4)  

  for {
    n, err = ro.Read(p)
    fmt.Println("n: ", n, err)

    if err != nil {
      if (err == io.EOF && n == 0 ){
        break 
      }
      fmt.Println("error reading Io: ", err)
    }

    fmt.Println(string(p[:n]))
    // subscribe to kube log 
    // watch for done 
  }
 
  

}
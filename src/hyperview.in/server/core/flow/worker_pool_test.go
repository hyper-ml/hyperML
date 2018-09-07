package flow_test


import(
  "fmt"
  "testing"
  .  "hyperview.in/server/core/flow"
)



func Test_WatchWorker(t *testing.T) {
  pk := NewWorkerPool()
  w_chan := NewWorkPoolWatcher()

  go pk.Watch(w_chan)

  for {
    select {
      case evt := <- w_chan:
        fmt.Println("Got event: ", evt)
    }
  }
}
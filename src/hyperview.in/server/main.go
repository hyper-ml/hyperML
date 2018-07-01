package main


import (
  "math/rand"
  "time"	
  "os"
	"os/signal"
	"syscall"
  "hyperview.in/server/base"
  "hyperview.in/server/rest"

)

func init() {
  rand.Seed(time.Now().UTC().UnixNano())
}

func main(){
  base.SetEnv()
  
  signalchannel := make(chan os.Signal, 1)
  signal.Notify(signalchannel, syscall.SIGHUP)

  go func() {
    for range signalchannel {
      base.Log("Handling SIGHUP signal.")
      //rest.HandleSighup()
    }
  }()

  rest.ServerMain()
}

  
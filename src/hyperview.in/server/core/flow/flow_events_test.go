package flow


/* commenting htis for now.
using DB events 
package flow

import (
  "testing"
  "fmt" 
  "time"
  zmq "github.com/pebbe/zmq4"
)


func Test_flowEventsWatcher(t *testing.T) { 
  flow_event_proc, err := NewEventProc("localhost", "5563")
  fmt.Println("event proc created: ", flow_event_proc, err)

  flow_event_proc.Watch()
  defer flow_event_proc.Close() 
}


func Test_flowEventsCreate(t *testing.T) { 

  //t1 := time.NewTimer(2*time.Second)
      
  fmt.Println("Test_flowEventsCreate")
  publisher, _ := zmq.NewSocket(zmq.PUB)
      
  publisher.Connect("tcp://localhost:5563")
  for {
  publisher.Send("$TERM$", 0) 
  publisher.Close() 
  }
}


func Test_sub(t *testing.T) {
  subscriber, _ := zmq.NewSocket(zmq.SUB)
  defer subscriber.Close()
  subscriber.Connect("tcp://localhost:5563")
  subscriber.SetSubscribe("A")

  for {
    //  Read envelope with address
    address, _ := subscriber.Recv(0)
    //  Read message contents
    //contents, _ := subscriber.Recv(0)
    fmt.Printf("[%s] \n", address)
  }
}

func Test_pub(t *testing.T) {
  publisher, _ := zmq.NewSocket(zmq.PUB)
  defer publisher.Close()
  publisher.Bind("tcp://*:5563")

  for {
    //  Write two messages, each with an envelope and content
    publisher.Send("FLOW", zmq.SNDMORE)
    publisher.Send("We don't want to see this", 0)
    publisher.Send("FLOW", zmq.SNDMORE)
    publisher.Send("We would like to see this", 0)
    publisher.Send("FLOW", zmq.SNDMORE)
    publisher.Send("$TERM$", 0)
    time.Sleep(time.Second)
  }
}
*/

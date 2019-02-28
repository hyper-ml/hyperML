package flow

/* Commenting this for now - Using DB level events 

package flow

import (
  "fmt" 
  "hyperflow.in/server/pkg/base"
  e "hyperflow.in/server/pkg/events"
  zmq "github.com/pebbe/zmq4"
)
 

const (
  FlowCreated e.EventType = iota
  FlowUpdated 
  FlowDeleted 
)
 


type flowEvents struct {
  ctx *zmq.Context
  socket *zmq.Socket
  ch chan *e.Event
  Done chan struct {}
}

func (fe *flowEvents) Ctx() *zmq.Context {
  return fe.ctx
}

func (fe *flowEvents) Receive() <- chan *e.Event {
  return fe.ch
}

func (fe *flowEvents) Close() {
  close(fe.Done)
  fe.socket.Close()
}


func NewEventProc(zmq_server string, receiver_port string) (*flowEvents, error) {
  
  //zctx, err := zmq.NewContext()
  sub, err := zmq.NewSocket(zmq.SUB)
  if err != nil {
    base.Log("[flow.NewEventProc] Failed to create a zmq socket or context: ", err)
    return nil, err
  }

  sub.Connect("tcp://" + zmq_server + ":" + receiver_port)
  sub.SetSubscribe("FLOW")



  ch := make(chan *e.Event)
  done := make(chan struct{})

  return &flowEvents {
    socket: sub,
    ch: ch,
    Done: done,
    //ctx: zctx,
  }, nil
}

func (fe *flowEvents) Watch() {
   
  fmt.Println("for loop starting")
  for {

    fmt.Println("start of for loop")
    addr, err := fe.socket.Recv(0)
    msg, err := fe.socket.Recv(0)
    fmt.Println("addr msg:", addr, msg)

    if err != nil {
      fmt.Println("error:", err)
      break
    }
    
    fmt.Println("recieved:", msg)
    if msg == "$TERM$" {
      break
    } 
    fmt.Println("End of for loop")
  } 
  fmt.Println("for loop exited")
  return 
} 


*/








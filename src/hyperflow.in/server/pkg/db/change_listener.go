package db
  
import(
  "fmt"
  "hyperflow.in/server/pkg/base"
)

type ChangeListener interface {

  RegisterObject(newch chan<- interface{}, objType interface{})

  Listen()

  // Register a new channel to receive broadcasts
  Register(chan<- interface{})

  // Unregister a channel so that it no longer receives broadcasts.
  Unregister(chan<- interface{})
  
  // Shut this broadcaster down.
  Close() error
  
  // Submit a new object to all subscribers
  TrackEvent(interface{}, interface{})
}


type changeListener struct {
  listening bool

  templates map[interface{}]bool
  input chan interface{}
  reg   chan chan<- interface{}
  unreg chan chan<- interface{} 

  outputs map[chan<- interface{}]bool
  done chan int
}



func (c *changeListener) broadcast(m interface{}) { 

  for ch := range c.outputs { 
    ch <- m
  }
}
 

func (c *changeListener) Listen() {
  
  if c.listening {
    base.Log("[ChangeListener.Listen] Already one listener is active")
    return 
  } 
  c.listening = true

  for { 
    select {
    case m := <-c.input:     
      c.broadcast(m)

    case ch, ok := <-c.reg:
      if ok { 
        c.outputs[ch] = true  
      } else { 
        break
      }
    case ch := <-c.unreg:
      delete(c.outputs, ch)

    case <-c.done:
      fmt.Println("quit")
      break
    } 
  } 
  return   
}


func NewChangeListener(buflen int) ChangeListener {
  c := &changeListener{
    templates: make(map[interface{}]bool),
    input:   make(chan interface{}, buflen),
    reg:     make(chan chan<- interface{}),
    unreg:   make(chan chan<- interface{}),
    outputs: make(map[chan<- interface{}]bool), 
    done: make(chan int),
  }

  go c.Listen()

  return c
}


func (c *changeListener) Register(newch chan<- interface{}) {
  c.reg <- newch
}

func (c *changeListener) RegisterObject(newch chan<- interface{}, objType interface{}) {

  if objType != nil {
    c.templates[objType] = true 
  }

  c.reg <- newch
}

func (c *changeListener) Unregister(newch chan<- interface{}) {
  c.unreg <- newch
}

func (c *changeListener) Close() error {
  close(c.done)
  return nil
}

func (c *changeListener) TrackEvent(m interface{}, template interface{}) {

  if c != nil {
    if c.templates[template] {
      c.input <- m
    }
  }

}
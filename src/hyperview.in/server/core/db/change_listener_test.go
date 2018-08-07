package db


import(
  "fmt" 
  "time"
  "sync" 
  "testing" 
)


func Test_ChangeListener(t *testing.T) {
  wg := sync.WaitGroup{}

  c := NewChangeListener(10)
  defer c.Close()
  
  
  for i := 0; i < 5; i++ {
   
    cch := make(chan interface{})

    c.Register(cch)
    fmt.Println("Registered channel:", i)

    go func() {
      defer wg.Done()
      defer c.Unregister(cch)
      <-cch
    }()
  }
  var dummy interface{}
  c.TrackEvent(1, dummy)

  wg.Wait()

}

func Test_ChangeMonitor(t *testing.T) {
  c := NewChangeListener(10)
  done:= make(chan struct{})
  var dummy interface{}
  listenOne(c, done)

  for i := 0; i < 5; i++ {
    fmt.Println("Sending in TestMonitor():", i)
    c.TrackEvent(i, dummy)
  }
  time.Sleep(2* time.Second)
  
  for i := 5; i < 10; i++ {
    fmt.Println("Sending in TestMonitor():", i)
    c.TrackEvent(i, dummy)
  } 
  fmt.Println("going to wait on done")
  <- done  

  fmt.Println("wait complete on done")
  c.Close()
}

func listenOne(c ChangeListener, done chan struct{}) { 

  ch := make(chan interface{})
  c.Register(ch)
  msg_count:=0  

  go func() { 
    for {
      select {
        case v := <- ch:
          fmt.Println("listenOne read: ", v)
          msg_count++
          if msg_count == 10 {
            fmt.Println("sending close on done")
            close(done)
          }
      }
    }   
    c.Unregister(ch)
  }() 
  
}

func Test_ChangeListenerCleanup(t *testing.T) {
  c := NewChangeListener(100)
  c.Register(make(chan interface{}))
  c.Close()
}


type fake_struct struct{}

func Test_DBContextListener(t *testing.T) {
  dbc, err := NewDatabaseContext(DB_NAME, DB_USER, DB_PASSWORD)
  if err != nil {
    t.Fatalf("Authentication Error in initialization of DB")
  }

  ch := make(chan interface{})
  done :=  make(chan struct{})

  var f fake_struct
  dbc.Listener.RegisterObject(ch, f) 
  
  go func() {
    for {
      select {
        case v := <- ch:
          fmt.Println("listenOne read: ", v)  
          close(done) 
          return
      } 
      fmt.Println("..")
    }
    fmt.Println("Exiting go routine")
  }()
  fmt.Println("in here before track event")
  dbc.Listener.TrackEvent(0, f)
  
  fmt.Println("in here after track event")
  <- done

  fmt.Println("in here after done")

  defer dbc.Listener.Unregister(ch) 
  fmt.Println("in here after Unregister")
  defer dbc.Close()
  fmt.Println("in here after Close")
}



func Test_NewListener(t *testing.T) {
  listener := NewChangeListener(10)
  //quit := make(chan int)

  //go listener.Listen()

  ch := make(chan interface{})
  d :=  make(chan struct{})

  var f fake_struct
  listener.RegisterObject(ch, f) 
  
  go func() {
    for {
      select {
        case v := <- ch:
          fmt.Println("listenOne read: ", v)  
          close(d) 
          return          
      }
    }
  }()


  listener.TrackEvent(0, f)
  <- d

  listener.Unregister(ch) 
  listener.Close()
  //quit <- 0
}



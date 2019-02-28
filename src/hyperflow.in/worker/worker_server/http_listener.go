package rest_server


import (
  "expvar"
  "net"
  "sync"
  "net/http"
  "time"
  "hyperflow.in/server/pkg/base"
)

//TODO: add exp var to capture stats
var httpListenerExpvars *expvar.Map
var maxWaitExpvar, maxActiveExpvar base.IntMax


func init() {
  httpListenerExpvars = expvar.NewMap("hyperflow_httpListener")
  httpListenerExpvars.Set("max_wait", &maxWaitExpvar)
  httpListenerExpvars.Set("max_active", &maxActiveExpvar)
}

func ListenAndServeHTTP(addr string, connLimit int,  readTimeout int, writeTimeout int, Handler http.Handler) error {
  listener, err := RatedListen("tcp", addr, connLimit)
  
  if err != nil {
    return err
  }

  defer listener.Close()
  server := &http.Server{Addr: addr, Handler: Handler}

  if readTimeout != 0 {
    server.ReadTimeout = time.Duration(readTimeout) * time.Second
  }
  
  if writeTimeout != 0 {
    server.WriteTimeout = time.Duration(writeTimeout) * time.Second
  }
  return server.Serve(listener)
}


type ratedListener struct {
  net.Listener
  active int
  limit int
  lock *sync.Cond
}

func RatedListen(protocol string, addr string, limit int) (net.Listener, error) {
  lner, err := net.Listen(protocol, addr)
  if err != nil || limit <= 0 {
    return lner, err
  }
  return &ratedListener{
    Listener: lner,
    limit: limit,
    lock: sync.NewCond(&sync.Mutex{}),
  }, nil
}

func (rl *ratedListener) Accept() (net.Conn, error) {
  conn, err := rl.Listener.Accept()

  // keep track of activeness and wait
  if err == nil {
    waitStart := time.Now()
    rl.lock.L.Lock()
    for rl.active >= rl.limit {
      rl.lock.Wait()
    }
    rl.active++
    maxActiveExpvar.SetIfMax(int64(rl.active))
    rl.lock.L.Unlock()
    waitTime := time.Since(waitStart)
    maxWaitExpvar.SetIfMax(int64(waitTime))
  }
  return &statConn{conn, rl}, err
}

func (rl *ratedListener) connServed() {
  rl.lock.L.Lock()
  rl.active--
  if rl.active == rl.limit-1 {
    rl.lock.Signal()
  }
  rl.lock.L.Unlock()
}

// conn object created to track closed requested 
type statConn struct {
  net.Conn
  listener *ratedListener
}

func (conn *statConn) Close() error {
  err := conn.Conn.Close()
  conn.listener.connServed()
  return err
}


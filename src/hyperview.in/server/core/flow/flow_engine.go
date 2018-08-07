package flow

import ( 
  //"time"
  //"context"
  //"hyperview.in/server/base/backoff"
  "hyperview.in/server/base"
  db_pkg "hyperview.in/server/core/db"
)

type FlowEngine interface{
  StartFlow(Id string) error
}


type flowEngine struct {
  qs *queryServer
  db *db_pkg.DatabaseContext
  namespace string
  defaultImage string
  dockerPullPolicy string

  // storage details - add later

}

func NewFlowEngine(qs *queryServer, db *db_pkg.DatabaseContext) *flowEngine{
  return &flowEngine{
    qs: qs,
    db: db,
  }
}

// monitor new messages from the worker pod or update on flow status
// end pods or mark flow completion 

func (fe *flowEngine) master(quit chan int) {
  
  eventCh := make(chan interface{})
  fe.db.Listener.RegisterObject(eventCh, Flow{})
  base.Log("[flowEngine.master] Starting Flow Master")
  for {
    select {
      case event_val, ok := <- eventCh:
        if !ok {
          return
        }

        FlowAttrs, ok := event_val.(*FlowAttrs)
        
        if !ok {
          base.Log("[flowEngine.master] Oops not a flow record")
          break
        } 

        base.Debug("[flowEngine.master] Flow Record: %v\n", FlowAttrs.Flow.Id)

      case <-quit:
        base.Log("[flowEngine.master] Quiting flow Engine master..")
        return
    }
  }

  return

  /*op := func() error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    eventCh := make(chan interface{})

    fe.db.Listener.RegisterObject(eventCh, Flow{})

    for {
      select {
        case val, ok := <- eventCh:
          base.Log("Received update:", val)
      }
    }

    return nil
  }

  backoff.RetryNotify(op, 
    backoff.NewExponentialBackOff(), 
    func(error, time.Duration) {
    })*/
}

 

func (fe *flowEngine) StartFlow(Id string) error {


  return nil
}
 

















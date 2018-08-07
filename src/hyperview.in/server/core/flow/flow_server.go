package flow

import ( 
  "hyperview.in/server/base"
  db_pkg "hyperview.in/server/core/db"
)

type FlowServer struct {
  db *db_pkg.DatabaseContext
  fe FlowEngine	
  qs *queryServer
  ns string
  quit chan int
}

func NewFlowServer(db *db_pkg.DatabaseContext, kubens string) *FlowServer {
    
  qs:= NewQueryServer(db)
  fe:= NewFlowEngine(qs, db)
  quit:= make(chan int)

  go fe.master(quit)
    
  fs := &FlowServer {
    db: db,
    ns: kubens,
    qs: qs,
    fe: fe,
    quit: quit,
  }

  return fs
    
}

func (fs *FlowServer) Close() {
  close(fs.quit)
}

func (fs *FlowServer) GetFlowAttr(flowId string) (*FlowAttrs, error) {
  return fs.qs.GetFlowAttr(flowId)
}


func (fs *FlowServer) RegisterWorker(flowId string, taskId string, ipaddr string) (*WorkerAttrs, error) {
  work_attr, err := fs.qs.registerW(flowId, taskId, ipaddr)
  
  if err != nil {
    base.Log("[FlowServer.RegisterWorker] Failed to register worker attributes: ", err)
    return nil, err
  }

  return work_attr, nil
}

func (fs *FlowServer) UnregisterWorker(workerId string) error {
  return fs.qs.unregisterW(workerId)
}





package rest


import (
  "fmt"
	"sync"
  "net/http"
  "time" 
  "hyperflow.in/server/pkg/flow"
  "hyperflow.in/server/pkg/auth"

  "hyperflow.in/server/pkg/storage"
  "hyperflow.in/server/pkg/db"
  ws "hyperflow.in/server/pkg/workspace"
  config_pkg "hyperflow.in/server/pkg/config"
)

type ServerContext struct {
	config *config_pkg.Config
	DBContext db.DatabaseContext
  
  lock sync.RWMutex
	statsTicker *time.Ticker
	HTTPClient  *http.Client
  
  authAPI auth.AuthServer
  objAPI storage.ObjectAPIServer
  logAPI storage.ObjectAPIServer
  flowAPI *flow.FlowServer

  wsAPI ws.ApiServer
//  vfsAPI *ws.VfsServer 
}


func NewServerContext(config *config_pkg.Config) (*ServerContext, error) {
 
  obj_api, err  := storage.NewObjectAPI(config.ObjStorage) 
  if err != nil { 
    return nil, fmt.Errorf("failed to initiate object server, err: ", err)
  }

  log_api, _ := storage.NewLogObjectAPI(config.ObjStorage)
  if err != nil { 
    return nil, fmt.Errorf("failed to initiate log server, err: ", err)
  }

  dbc, err := db.NewDatabaseContext(config.DB) 
  if err != nil {
    return nil, fmt.Errorf("failed to initiate db, err: ", err)
  }

  ws_api, err := ws.NewApiServer(dbc, obj_api)
  if err != nil { 
    return nil, fmt.Errorf("failed to iniate workspace server, err: ", err)
  }

  //vfs := ws.NewVfsServer(dbc, obj_api) 
  flow_server, err := flow.NewFlowServer(config, dbc, obj_api, ws_api, log_api)
  if err != nil {
    return nil, fmt.Errorf("failed to initate flow server, err: ", err)
  }

  auth_server := auth.NewAuthServer(dbc)

  sc := &ServerContext{
    config: config,
    DBContext: dbc,
    HTTPClient: http.DefaultClient,
    objAPI: obj_api,
    logAPI: log_api,
    wsAPI: ws_api,
    //vfsAPI: vfs, 
    flowAPI: flow_server,
    authAPI: auth_server,
  }
  return sc, nil
}


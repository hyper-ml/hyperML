package rest


import (
	"sync"
  "net/http"
  "time"
  "hyperview.in/server/core/tasks"
  "hyperview.in/server/core/flow"

  "hyperview.in/server/core/storage"
  "hyperview.in/server/core/db"
  "hyperview.in/server/core/workspace"
  "hyperview.in/server/base"
)

const (
  LOG_DIR = "logs"
  DIR_SEPARATOR = "/"
)

type ServerContext struct {
	config *ServerConfig
	lock sync.RWMutex
	statsTicker *time.Ticker
	HTTPClient  *http.Client
  objectAPI storage.StorageServer
  logApi storage.StorageServer
  databaseContext *db.DatabaseContext
  workspaceApi workspace.ApiServer
  vfs *workspace.VfsServer
  tasker tasks.Tasker
  flowServer *flow.FlowServer
}

 

// TODO: add error logging and response
// accessed by HTTP Handlers so needs to be thread safe 
func NewServerContext(config *ServerConfig) *ServerContext{

  dir := config.BaseDir
  log_dir := config.BaseDir + "/" + LOG_DIR

  //TODO: add config variable for storage option
  oapi, err  := storage.NewObjectAPI(dir, 0, storage.GoogleStorage) 
  if err != nil {
    base.Error("[NewServerContext] object API  Error: ", err)
    return nil
  }

  log_api, _ := storage.NewObjectAPI(log_dir, 0, storage.GoogleStorage)
  if err != nil {
    base.Error("[NewServerContext] log API  Error: ", err)
    return nil
  }

  dbc, err := db.NewDatabaseContext(config.DatabaseConfig.Name, config.DatabaseConfig.User, config.DatabaseConfig.Pass) 
  if err != nil {
    base.Error("[NewServerContext] DB Context Error: ", err)
    return nil
  }

  ws_api, err := workspace.NewApiServer(dbc, oapi)
  if err != nil {
    base.Error("[NewServerContext] Ws API Error: ", err)
    return nil
  }

  vfs := workspace.NewVfsServer(dbc, oapi)
  tasker := tasks.NewShellTasker(dbc)

  kube_namespace:= "hflow"
  flow_server := flow.NewFlowServer(dbc, kube_namespace, oapi, ws_api)

  sc := &ServerContext{
    config: config,
    HTTPClient: http.DefaultClient,
    objectAPI: oapi,
    logApi: log_api,
    workspaceApi: ws_api,
    databaseContext: dbc,
    vfs: vfs,
    tasker: tasker,
    flowServer: flow_server,
  }
  return sc
}


func (sc *ServerContext) getBasePath() string {
  return sc.config.BaseDir
}

func (sc *ServerContext) getLogBasePath() string {
  return sc.config.BaseDir + DIR_SEPARATOR + LOG_DIR
}
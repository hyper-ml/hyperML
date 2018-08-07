package rest


import (
	"sync"
  "net/http"
  "time"
  "strconv"
  "hyperview.in/server/core/tasks"
  "hyperview.in/server/core/flow"

  "hyperview.in/server/core/storage"
  "hyperview.in/server/core/db"
  "hyperview.in/server/core/workspace"
  "hyperview.in/server/base"

)

type ServerContext struct {
	config *ServerConfig
	lock sync.RWMutex
	statsTicker *time.Ticker
	HTTPClient  *http.Client
  objectAPI storage.StorageServer
  databaseContext *db.DatabaseContext
  workspaceApi workspace.ApiServer
  vfs *workspace.VfsServer
  tasker tasks.Tasker
  flowServer *flow.FlowServer
}

// TODO: add error logging and response
// accessed by HTTP handlers so needs to be thread safe 
func NewServerContext(config *ServerConfig) *ServerContext{

  dir := config.BaseDir

  cacheStr := base.GetEnv("OBJECT_CACHE_BYTES")
  var cacheBytes int64
  var err error

  if (cacheStr != "") {
    cacheBytes, err = strconv.ParseInt(cacheStr, 10, 64)
    if err == nil {
      base.Log("invalid cache bytes.")
    }
  }


  //TODO: add config variable for storage option
  oapi, _ := storage.NewObjectAPI(dir, cacheBytes, "GCS") 
  dbc, err := db.NewDatabaseContext(config.DatabaseConfig.Name, config.DatabaseConfig.User, config.DatabaseConfig.Pass) 
  ws_api, err := workspace.NewApiServer(dbc, oapi)
  vfs := workspace.NewVfsServer(dbc, oapi)
  tasker := tasks.NewShellTasker(dbc)

  kube_namespace:= "hflow"
  flow_server := flow.NewFlowServer(dbc, kube_namespace)

  sc := &ServerContext{
    config: config,
    HTTPClient: http.DefaultClient,
    objectAPI: oapi,
    workspaceApi: ws_api,
    databaseContext: dbc,
    vfs: vfs,
    tasker: tasker,
    flowServer: flow_server,
  }
  return sc
}
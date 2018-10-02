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
  STORAGE_BASE_DIR = "hyperview"
  LOG_DIR = "logs"
  SEP = "/"

)

type ServerContext struct {
	config *RunConfig
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

func NewServerContext(c *RunConfig) *ServerContext{

  storage_dir := STORAGE_BASE_DIR
  log_dir := storage_dir + SEP + LOG_DIR 

  oapi, err  := storage.NewObjectAPI(storage_dir, c.StorageOption) 
  if err != nil {
    base.Error("[NewServerContext] object API  Error: ", err)
    return nil
  }

  log_api, _ := storage.NewObjectAPI(log_dir, c.StorageOption)
  if err != nil {
    base.Error("[NewServerContext] log API  Error: ", err)
    return nil
  }

  dbc, err := db.NewDatabaseContext(c.driver, c.DatabaseConfig.DbName, c.DbConfig.Dbuser, c.DbConfig.Dbpass) 
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

  flow_server := flow.NewFlowServer(dbc, oapi, ws_api)

  sc := &ServerContext{
    config: c,
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
  return STORAGE_BASE_DIR
}

func (sc *ServerContext) getLogBasePath() string {
  return STORAGE_BASE_DIR + SEP + LOG_DIR
}
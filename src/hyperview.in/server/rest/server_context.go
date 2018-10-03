package rest


import (
	"sync"
  "net/http"
  "time" 
  "hyperview.in/server/core/flow"

  "hyperview.in/server/core/storage"
  "hyperview.in/server/core/db"
  "hyperview.in/server/core/workspace"
  "hyperview.in/server/base"
  config_pkg "hyperview.in/server/config"
)

const (

  GCS = "GCS"
  S3 = "S3"
)

type ServerContext struct {
	config *RunConfig
	lock sync.RWMutex
	statsTicker *time.Ticker
	HTTPClient  *http.Client
  objectAPI storage.StorageServer
  logger storage.StorageServer
  databaseContext *db.DatabaseContext
  workspaceApi workspace.ApiServer
  vfs *workspace.VfsServer 
  flowServer *flow.FlowServer
}

func setGCSVars(gc *config_pkg.GcsConfig) error {
  if gc != nil {
    return base.SetEnvVar("GOOGLE_APPLICATION_CREDENTIALS", gc.CredPath)
  }
  return nil
}
 

func NewServerContext(c *RunConfig) (*ServerContext, error) {

  if c.Current.StorageOption == config_pkg.GStorage {
    if err := setGCSVars(c.Current.Gcs); err != nil {
    return nil, err
    }
  }

  oapi, err  := storage.NewObjectAPI(c.Current.StorageOption, c.StorageBaseDir, c.Current.S3, c.Current.Gcs) 
  if err != nil {
    base.Error("[NewServerContext] object API  Error: ", err)
    return nil, err
  }

  logger, _ := storage.NewObjectAPI(c.Current.StorageOption, c.LogBaseDir, c.Current.S3, c.Current.Gcs)
  if err != nil {
    base.Error("[NewServerContext] log API  Error: ", err)
    return nil, err
  }
  
  dbc, err := db.NewDatabaseContext(c.Current.DbConfig.Driver, c.Current.DbConfig.DbName, c.Current.DbConfig.Dbuser, c.Current.DbConfig.Dbpass) 
  if err != nil {
    base.Error("[NewServerContext] DB Context Error: ", err)
    return nil, err
  }

  ws_api, err := workspace.NewApiServer(dbc, oapi)
  if err != nil {
    base.Error("[NewServerContext] Ws API Error: ", err)
    return nil, err
  }

  vfs := workspace.NewVfsServer(dbc, oapi) 

  flow_server := flow.NewFlowServer(dbc, oapi, ws_api, logger)

  sc := &ServerContext{
    config: c,
    HTTPClient: http.DefaultClient,
    objectAPI: oapi,
    logger: logger,
    workspaceApi: ws_api,
    databaseContext: dbc,
    vfs: vfs, 
    flowServer: flow_server,
  }
  return sc, nil
}


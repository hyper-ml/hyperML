package rest

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/auth"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/flow"
	"net/http"
	"sync"
	"time"

	constants "github.com/hyper-ml/hyperml/server/pkg/config"
	"github.com/hyper-ml/hyperml/server/pkg/db"
	"github.com/hyper-ml/hyperml/server/pkg/pods"
	qs_pkg "github.com/hyper-ml/hyperml/server/pkg/qs"
	"github.com/hyper-ml/hyperml/server/pkg/requests"
	"github.com/hyper-ml/hyperml/server/pkg/schedules"
	"github.com/hyper-ml/hyperml/server/pkg/storage"
	ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
)

// ServerContext : Retains Server Context Parameters
type ServerContext struct {
	config    *constants.Config
	DBContext db.DatabaseContext

	lock        sync.RWMutex
	statsTicker *time.Ticker
	HTTPClient  *http.Client

	authAPI auth.Server
	objAPI  storage.ObjectAPIServer
	logAPI  storage.ObjectAPIServer
	flowAPI *flow.FlowServer

	wsAPI ws.ApiServer
	//  vfsAPI *ws.VfsServer

	podkeeper *pods.Keeper
	usrReq    *requests.RequestHandler
	qs        *qs_pkg.QueryServer

	scheduler *schedules.NotebookScheduler

	DisablePodRequest bool
}

// NewServerContext : Initiates a new server context
func NewServerContext(config *constants.Config) (*ServerContext, error) {
	var err error
	sc := &ServerContext{
		config: config,
	}

	if config.GetBool(constants.Safemode) {
		base.Out("Starting in safemode. k8s Errors will be ignored ")
	}

	sc.DBContext, err = db.NewDatabaseContext(config.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate db, err: %v", err)
	}

	// initiate query server
	sc.qs = qs_pkg.NewQueryServer(sc.DBContext)

	sc.authAPI = auth.NewAuthServer(sc.qs, config.NoAuth)

	if !config.GetBool("DisableFlow") {
		sc.objAPI, err = storage.NewObjectAPI(config.ObjStorage)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate object server, err: %v", err)
		}

		sc.logAPI, err = storage.NewLogObjectAPI(config.ObjStorage)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate log server, err: %v", err)
		}

		sc.wsAPI, err = ws.NewApiServer(sc.DBContext, sc.objAPI)
		if err != nil {
			return nil, fmt.Errorf("failed to iniate workspace server, err: %v", err)
		}

		sc.flowAPI, err = flow.NewFlowServer(config, sc.DBContext, sc.objAPI, sc.wsAPI, sc.logAPI)
		if err != nil {
			return nil, fmt.Errorf("failed to initate flow server, err: %v", err)
		}
	}

	// initiate pod keeper
	sc.podkeeper, err = pods.NewKeeper(config, sc.qs)

	if err != nil {

		if !config.GetBool(constants.Safemode) {
			return sc, err
		}

		base.Error("Failed to Initiate Kubernetes Connection:", err)
		base.Error("As the safe mode is ON, the server will allow queries but new Requests will now be disabled")
		sc.DisablePodRequest = true
	}
	sc.scheduler = schedules.NewNotebookScheduler(sc.qs, sc.podkeeper, config)
	sc.usrReq = requests.NewRequestHandler(sc.qs, config, sc.podkeeper, sc.scheduler)

	return sc, nil
}

// ReadyForRequests : Disable pod requests in Safe Mode
func (sc *ServerContext) ReadyForRequests() bool {
	return !sc.DisablePodRequest
}

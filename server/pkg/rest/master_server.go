package rest

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/config"
	"net/http"
)

// DefaultMaxIncomingConnections : No limit by default
var DefaultMaxIncomingConnections = 0

// ServerReadTimeout : Timeout limit on read requests
var ServerReadTimeout = 200

// ServerWriteTimeout : Timeout limit on write requests
var ServerWriteTimeout = 200

var master *Master

// Master : Master Server that handlers all REST calls
type Master struct {
	Config *config.Config
	IP     string
	Port   int32
	SSL    bool
}

// NewMasterServer : Handles creation of a new Master
func NewMasterServer(ip string, port int32, ssl bool, c *config.Config) *Master {
	return &Master{
		Config: c,
		IP:     ip,
		Port:   port,
		SSL:    ssl,
	}
}

// ListenAddress : Returns the listening address for the server from config
func (m *Master) ListenAddress() string {
	var addr string
	if m.IP != "" || m.Port != 0 {
		// if !m.SSL {
		//	addr = "http"
		//} else {
		//	addr = "https"
		//}

		//addr = addr + "://"

		if m.IP != "" {
			addr = addr + m.IP
		} else {
			addr = addr + "0.0.0.0"
		}

		if m.Port != 0 {
			addr = addr + ":" + fmt.Sprintf("%d", m.Port)
		} else {
			addr = addr + ":" + fmt.Sprintf("%d", config.DefaultMasterPort)
		}
		return addr
	}

	return m.Config.GetListenAddr()
}

// Serve : Serves the Master
func (m *Master) Serve(Handler http.Handler) {
	err := ListenAndServeHTTP(m.ListenAddress(), DefaultMaxIncomingConnections, ServerReadTimeout, ServerWriteTimeout, Handler)

	if err != nil {
		base.Log("failed to start HTTP Server on %s: %v", m.ListenAddress(), err)
	}
}

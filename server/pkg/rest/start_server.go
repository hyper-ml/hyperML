package rest

import (
	"github.com/hyper-ml/hyperml/server/pkg/base"
	hfConfig "github.com/hyper-ml/hyperml/server/pkg/config"
)

// StartServer : Initiates Master server afte injecting config params
func StartServer(listenIP string, listenPort int, isSSL bool, configPath string) {
	// todo: ssl
	c, err := hfConfig.NewConfig(listenIP, listenPort, configPath)
	if err != nil {
		raisePanic(err)
	}

	m := NewMasterServer(listenIP, int32(listenPort), isSSL, c)
	runServer(m)
}

func runServer(m *Master) {

	sc, err := NewServerContext(m.Config)
	if err != nil {
		raisePanic(err)
	}

	base.Println("Starting server on %s ...", m.ListenAddress())
	m.Serve(CreatePublicHandler(sc))
}

func raisePanic(err error) {
	panic(err)
}

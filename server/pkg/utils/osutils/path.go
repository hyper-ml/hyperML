package osutils

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"os"
	"path/filepath"
)

// AppHome : Application Home directory. usually ~/.hyperml
func AppHome() string {
	userHome, _ := base.HomeDir()
	hfPath := filepath.Join(userHome, AppHomePath)
	os.MkdirAll(hfPath, DefaultPerm)
	return hfPath
}

// ServerHome : Returns or Creates server home path. usually ~/.hyperml/server
func ServerHome() string {
	appPath := AppHome()
	serverPath := filepath.Join(appPath, ServerHomePath)

	os.MkdirAll(serverPath, os.ModePerm)
	return serverPath
}

// ClientHome : Returns or Creates server home path. usually usually ~/.hyperml/client
func ClientHome() string {
	appPath := AppHome()
	clientPath := filepath.Join(appPath, ClientHomePath)
	os.MkdirAll(clientPath, os.ModePerm)
	return clientPath
}

// CreateDataDirPath : Appends Server home directory with data directory path
// creates any missing directories
func CreateDataDirPath(datapath string) string {
	dataDirPath := filepath.Join(ServerHome(), datapath)

	if err := os.MkdirAll(dataDirPath, os.ModePerm); err != nil {
		base.Error(err.Error())
		panic(err)
	}

	return dataDirPath
}

// ConfigDefaultPath : Returns server config file path
func ConfigDefaultPath() string {
	return filepath.Join(ServerHome(), ConfigFileName)
}

// ConfigDefaultPathValid : Returns server config file path if file exists
func ConfigDefaultPathValid() string {
	p := filepath.Join(ServerHome(), ConfigFileName)

	if PathExists(p) {
		return p
	}

	return nullString
}

// K8ConfigPath :
func K8ConfigPath() string {
	home, _ := base.HomeDir()
	return filepath.Join(home, ".kube", "config")
}

// K8ConfigValidPath :
func K8ConfigValidPath() (p string, fnerr error) {
	p = K8ConfigPath()

	if p != "" {
		if PathExists(p) {
			return p, nil
		}
		fnerr = fmt.Errorf("No valid kubernetes config found: %v", p)
		return
	}

	return p, fmt.Errorf("Kubernetes config not found in $HOME dir")
}

// PathExists : Checks if a file exists at a given path
func PathExists(p string) bool {

	if p == nullString {
		return false
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

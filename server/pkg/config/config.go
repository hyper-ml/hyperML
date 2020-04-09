package config

import (
	"encoding/json"
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/utils/osutils"
	"io/ioutil"
	"os"
	"strings"
)

// Config : holds entire configuration of hflow server
type Config struct {
	Namespace          string
	FlowNamespace      string
	PublicInterface    string
	MasterIP           string
	MasterPort         int32
	MasterExternalPort int32
	NoSSL              bool
	NoAuth             bool
	DisableFlow        bool

	DB          *DBConfig
	K8          *KubeConfig
	ObjStorage  *ObjStorageConfig
	Notebooks   *NotebookConfig
	NbScheduler *NbSchedulerConfig

	Jobs *JobConfig
	Pods *PodConfig

	Domain   string
	LogLevel int
	LogPath  string
	Safemode bool
}

// GetDomain : Get domain used to generate URL for launch notebooks/labs
func (c *Config) GetDomain() string {
	if c.Domain == "" {
		c.Domain = DefaultDomain
	}
	return c.Domain
}

// GetNS : Get namespace for kuberentes cluster
func (c *Config) GetNS() string {
	if c.Namespace == "" {
		c.Namespace = DefaultNamespace
	}
	return c.Namespace
}

// GetFlowNS : Get kubernetes namespace to be used by Flow
func (c *Config) GetFlowNS() string {
	if c.FlowNamespace == "" {
		c.FlowNamespace = c.GetNS()
	}

	return c.FlowNamespace
}

// NewConfig : Used to fire up new configuration object
func NewConfig(listenIP string, listenPort int, p string) (*Config, error) {
	var listenPort32 int32
	listenPort32 = int32(listenPort)

	if p != "" {
		fmt.Println("Reading config from path: ", p)
		return NewConfigFromPath(p)
	}

	// check vars are set directly in OS environment
	if envVar := osutils.GetOsEnvVar(ConfigVarsEnvName); envVar != "" {
		fmt.Println("Reading config from Environment Var: ", envVar)
		return NewConfigFromJSON(envVar)
	}

	// check if config path is set in environtment variable
	if cPath := osutils.GetOsEnvVar(ConfigPathEnvName); cPath != "" {
		fmt.Println("Reading config from Env Var path: ", cPath)
		return NewConfigFromPath(cPath)
	}

	// check default validated path
	if d := osutils.ConfigDefaultPathValid(); d != "" {
		fmt.Println("Reading config from default HFLOW path: ", d)
		return NewConfigFromPath(d)
	}

	fmt.Println("No Config files found. Generating default config...")
	return NewDefaultConfig(listenIP, listenPort32)
}

// GetListenAddr : Get HTTP listener address for main server
func (c *Config) GetListenAddr() string {
	portStr := fmt.Sprint(c.MasterPort)
	return c.MasterIP + ":" + portStr
}

// GetNotebooks : Returns Notebook Config if set. else Default
func (c *Config) GetNotebooks() *NotebookConfig {
	if c.Notebooks != nil {
		return c.Notebooks
	}

	return &NotebookConfig{}
}

// GetNbScheduler : Returns Notebook Scheduler Config
func (c *Config) GetNbScheduler() *NbSchedulerConfig {
	if c.NbScheduler != nil {
		return c.NbScheduler
	}
	return &NbSchedulerConfig{}
}

// GetJobs : Get Job specific config from default object
func (c *Config) GetJobs() *JobConfig {
	if c.Jobs != nil {
		return c.Jobs
	}
	return &JobConfig{
		BackoffLimit:    JobDefaultBackoff,
		DeadlineSeconds: JobDefaultDeadline,
	}
}

// Get : Get string config vars
func (c *Config) Get(vname string) (string, error) {
	switch vname {
	case "MasterIP":
		return c.MasterIP, nil
	}
	return "", fmt.Errorf("Invalid Config Param")
}

// GetInt32 : Get Int32 config vars
func (c *Config) GetInt32(vname string) (int32, error) {
	switch vname {
	case "MasterPort":
		return c.MasterPort, nil
	case "MasterExternalPort":
		return c.MasterExternalPort, nil
	}
	return 0, fmt.Errorf("Invalid Config Param")
}

// GetBool : Get boolean config vars
func (c *Config) GetBool(vname string) bool {
	switch vname {
	case "DisableFlow":
		return c.DisableFlow
	case Safemode:
		return true
		//return c.Safemode
	}
	return false
}

// NewDefaultConfig : Default Config used when stored config is not found
func NewDefaultConfig(masterIP string, masterPort int32) (*Config, error) {
	var err error
	k8configpath, _ := osutils.K8ConfigValidPath()
	datafiles := osutils.CreateDataDirPath(DefaultDataDirPath)

	c := &Config{
		NoSSL:    true,
		NoAuth:   true,
		Safemode: true,
		LogLevel: 5,
		DB: &DBConfig{
			Driver:      Badger,
			Name:        DefaultDBBucket,
			DataDirPath: datafiles,
			EventBuffer: DefaultDBEventBuffer,
		},
		K8: &KubeConfig{
			InCluster: false,
			Path:      k8configpath,
		},
		Domain:      DefaultDomain,
		DisableFlow: true,
		Notebooks: &NotebookConfig{
			IP:            DefaultNotebookPort,
			Port:          DefaultNotebookIP,
			BasePath:      DefaultNotebookBasePath,
			Command:       NbDefaultCommand,
			BackgroundCmd: NbDefaultBackgroundCmd,
		},
		NbScheduler: &NbSchedulerConfig{
			Concurrency:   0,
			SweepInterval: NbDefaultSweepInterval, // Minutes
			SaveInterval:  NbDefaultSaveInterval,  // Seconds
		},
		Jobs: &JobConfig{
			BackoffLimit:    JobDefaultBackoff,
			DeadlineSeconds: JobDefaultDeadline,
		},
		Pods: &PodConfig{},
	}

	if masterIP != "" {
		c.MasterIP = masterIP
	}

	if masterPort != 0 {
		c.MasterPort = masterPort
	} else {
		c.MasterPort = DefaultMasterPort
	}

	if c.PublicInterface != "" {
		proto := "http://"
		if !c.NoSSL {
			proto = "https://"
		}

		c.PublicInterface = proto + masterIP + ":" + string(masterPort)
	}

	c, err = SaveConfig(c, "")
	if err != nil {
		fmt.Println("Failed to write default config at default path: %", err)
	}
	return c, nil
}

// NewConfigFromPath : Used when a config path is set and file exists
func NewConfigFromPath(p string) (*Config, error) {
	var c *Config
	var err error

	if !osutils.PathExists(p) {
		return nil, fmt.Errorf("config path does not exist: %v", p)
	}

	jsonData, err := ioutil.ReadFile(p)
	if err != nil {
		base.Error("failed to read config from "+p+" , err: ", err)
		return nil, err
	}

	err = json.Unmarshal(jsonData, &c)
	if err != nil {
		base.Error("failed to understand config json from given path, err: ", err)
		base.Error("json value: ", jsonData)
		return nil, err
	}

	return c, nil
}

// NewConfigFromJSON : Sets config file with json string input
func NewConfigFromJSON(cString string) (c *Config, fnerr error) {

	fnerr = json.Unmarshal([]byte(cString), &c)
	if fnerr != nil {
		base.Error("failed to unmarshal config string from OS env, err: ", cString, fnerr)
		return
	}

	return c, nil
}

// SaveConfig : Saved config file to a give path (default $HOME/.hflow)
func SaveConfig(c *Config, p string) (*Config, error) {
	var targetPath string

	if p != "" {
		targetPath = p
	} else {
		targetPath = osutils.ConfigDefaultPath()
	}

	jsonData, err := json.Marshal(c)
	if err != nil {
		base.Error("failed to write config to default path "+targetPath+", err:", err)
		return nil, err
	}

	err = ioutil.WriteFile(targetPath, jsonData, os.FileMode(ConfigFilePerm))
	if err != nil {
		base.Error("Failed to create default config file at "+p+", err:", err)
		return nil, err
	}

	return c, nil

}

// DbTarget : Identifies database option (postgres, badger, bolt etc)
type DbTarget string

// DBConfig : Stores config for metadata DB
type DBConfig struct {
	Driver      DbTarget // POSTGRES, BOLT
	Name        string
	User        string
	Pass        string
	DataDirPath string
	//Change Listener Threshold
	EventBuffer int
}

// KubeConfig : Stores K8S connection info
type KubeConfig struct {
	Namespace string
	Path      string
	InCluster bool
}

// StorageTarget : Identifies target backend for object storage
type StorageTarget string

// ObjStorageConfig : Stored Object Storage backend config
type ObjStorageConfig struct {
	StorageTarget StorageTarget
	BaseDir       string
	S3            *S3Config
	Gcs           *GcsConfig
}

// S3Config : Stored Object Storage S3 backend config
type S3Config struct {
	CredPath     string
	AccessKey    string
	SecretKey    string
	SessionToken string
	Bucket       string
	Region       string
	Creds        string
}

// GcsConfig : Stored Object Storage GCS backend config
type GcsConfig struct {
	CredsPath string
	Bucket    string
	Creds     []byte
}

// NotebookConfig : Configuration Parameters for Notebook instance
type NotebookConfig struct {
	Command       string
	Port          string
	IP            string
	BasePath      string
	BackgroundCmd string
}

// GetCommand : Returns Default command to launch notebook
func (nbc *NotebookConfig) GetCommand() string {
	tokenMap := strings.NewReplacer("{port}", nbc.Port, "{ip}", nbc.IP)
	return tokenMap.Replace(nbc.Command)
}

// GetBackgroundCmd : Returns default command to launch background notebook
func (nbc *NotebookConfig) GetBackgroundCmd() string {
	if nbc != nil {
		return nbc.BackgroundCmd
	}
	return NbDefaultBackgroundCmd
}

// NbSchedulerConfig : Configuration Parameters for Notebook Scheduler
type NbSchedulerConfig struct {
	SweepInterval int
	SaveInterval  int
	Concurrency   int
}

// JobConfig : Job specific config
type JobConfig struct {
	DeadlineSeconds int
	BackoffLimit    int
	TTL             int32
}

// PodConfig : Job specific config
type PodConfig struct {
	TTL int32
}

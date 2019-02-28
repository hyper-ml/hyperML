package config

import (
  "os"
  "fmt"
  "strconv"
  "io/ioutil"
  "encoding/json"
 
  "hyperflow.in/server/pkg/base"
  "hyperflow.in/server/pkg/utils/os_utils"
)
 
const (
  ConfigVarsEnvName = "HF_SERVER_CONFIG_VARS"
  ConfigPathEnvName ="HF_SERVER_CONFIG_PATH"
  GCS StorageTarget = "GCS"
  S3 StorageTarget = "S3"
  Bolt DbTarget = "BOLT"
  Postgres DbTarget = "POSTGRES"
  Badger DbTarget = "BADGER"

)
 
type StorageTarget string
type DbTarget string

var (
  DefaultDataDirPath = "/var/tmp/hflow.db"
  DefaultDBBucket = "hflow_master"
  DefaultStorageBucket = "hyperflow"
  DefaultStorageTarget = GCS
  DefaultStorageBaseDir = "hyperflow"
  DefaultMasterPort = int32(8888)
  ConfigFilePerm = 0644
)

type Config struct {
  PublicInterface string
  MasterIp string
  MasterPort int32
  MasterExternalPort int32
  NoSSL bool

  DB *DBConfig
  K8 *KubeConfig
  ObjStorage *ObjStorageConfig 

  LogLevel string
  LogPath string
}
 
func NewConfig(listenIp, listenPort, p string) (*Config, error) { 
  var listen_port int32

  if listenPort != "" {
    port64, err := strconv.ParseInt(listenPort, 10, 32)
    if err != nil {
      return nil, fmt.Errorf("failed to parse listen port %s, err: %v", listenPort, err)
    }
    listen_port = int32(port64)
  }

  if p != "" {
    return NewConfigFromPath(p)  
  } 

  // check env
  if c_string := os_utils.GetOsEnvVar(ConfigVarsEnvName); c_string != "" { 
    return NewConfigFromJson(c_string)
  }

  if c_path := os_utils.GetOsEnvVar(ConfigPathEnvName); c_path != "" {
    return NewConfigFromPath(c_path)
  }

  // check default validated path 
  if d := os_utils.HFConfigDefaultPathV(); d != "" {
    return NewConfigFromPath(d)
  } 

  return NewDefaultConfig(listenIp, listen_port)    
}

func (c *Config) GetListenAddr() string {
  port_string := fmt.Sprint(c.MasterPort)
  return c.MasterIp + ":" + port_string
}

func NewDefaultConfig(masterIp string, masterPort int32) (*Config, error) {

  c := &Config{
    NoSSL: true,
    LogLevel: "5",
    DB: &DBConfig{
      Driver: Bolt, 
      Name: DefaultDBBucket, 
      DataDirPath: DefaultDataDirPath,
    },
    K8: &KubeConfig{
      InCluster: true,
    },
    ObjStorage: &ObjStorageConfig{
      StorageTarget: DefaultStorageTarget,
      BaseDir: DefaultStorageBaseDir,   
      Gcs: &GcsConfig{
        Bucket: DefaultStorageBucket,
      },
      S3: &S3Config{
        Bucket: DefaultStorageBucket,
        Region: "us-west-2",
      },
    }, 
  }

  if masterIp != "" {
    c.MasterIp = masterIp
  }

  if masterPort != 0 {
    c.MasterPort = masterPort
  } else {
    c.MasterPort = DefaultMasterPort
  }

  if c.PublicInterface != "" {
    proto := "http://"
    if !c.NoSSL{
      proto = "https://"
    } 

    c.PublicInterface = proto + masterIp + ":" + string(masterPort)
  }

  return c, nil
}

func NewConfigFromPath(p string) (*Config, error) {
  var c *Config
  var err error 

  if !os_utils.PathExists(p) {
    return nil, fmt.Errorf("config path does not exist: ", p)
  } 

  j_string, err := ioutil.ReadFile(p)
  if err != nil {
    base.Error("failed to read config from " + p + " , err: ", err)
    return nil, err
  }

  err = json.Unmarshal(j_string, &c)
  if err != nil {
    base.Error("failed to understand config json from given path, err: ", err)
    base.Error("json value: ", j_string)
    return nil, err
  }

  return c, nil
}

func NewConfigFromJson(cString string) (c *Config, fnerr error) {
  
  fnerr = json.Unmarshal([]byte(cString), &c)
  if fnerr != nil {
      base.Error("failed to unmarshal config string from OS env, err: ", cString, fnerr)
      return
  }

  return c, nil
}
 

func SaveConfig(c *Config, p string) (*Config, error) {
  var target_path string

  if p != "" {
    target_path = p
  } else {
    target_path = os_utils.HFConfigDefaultPath()
  }

  j_string, err := json.Marshal(c)
  if err != nil {
    base.Error("failed to write config to default path " + target_path + ", err:", err)
    return nil, err
  }

  err = ioutil.WriteFile(target_path, j_string, os.FileMode(ConfigFilePerm))
  if err != nil {
    base.Error("Failed to create default config file at " + p +", err:", err)
    return nil, err
  }

  return c, nil

} 


type DBConfig struct {
  Driver  DbTarget // POSTGRES, BOLT 
  Name string
  User string
  Pass string
  DataDirPath string
}

type KubeConfig struct {
  Namespace string
  Path string 
  InCluster bool
}

type ObjStorageConfig struct {
  StorageTarget StorageTarget
  BaseDir string
  S3 *S3Config
  Gcs *GcsConfig
}

type S3Config struct {
  CredPath string
  AccessKey string
  SecretKey string
  SessionToken string
  Bucket string
  Region string
  Creds string
}

type GcsConfig struct {
  CredsPath string
  Bucket string
  Creds []byte
}

package main



import(
  "os"
  "github.com/spf13/cobra"

  "hyperflow.in/server/pkg/base"
  "hyperflow.in/server/pkg/rest"
  "hyperflow.in/server/pkg/config"
)


func RootCmd() (*cobra.Command, error) {
  var ip string
  var port string
  var config_path string 

  root_cmd := &cobra.Command{
    Use: os.Args[0],
    Short: "hyperflow Server Command line",
    Long: `hyperflow Server Command line`,
    Run: func(cmd *cobra.Command, args []string) {
      base.Info("Starting Server... ")

      rest.StartServer(ip, port, config_path)      
    },
  }

  root_cmd.PersistentFlags().StringVarP(&ip, "listen-ip", "", "", "Listen IP with port. Defaults to localhost")
  root_cmd.PersistentFlags().StringVarP(&port, "listen-port", "", "", "Listen Port. Defaults to 8888.")
  root_cmd.PersistentFlags().StringVarP(&config_path, "config", "c", "", "Path to config file. Defaults to $HOME/.hflow")  

  root_cmd.AddCommand(configCmd())
  root_cmd.AddCommand(authCommand())
  
  task_cmds:= taskCmds()
  for _, cmd := range task_cmds {
    root_cmd.AddCommand(cmd)
  }

  return root_cmd, nil
}

func configCmd() (*cobra.Command) {
  var config_path string
  var storage_opt string

  var s3_creds string
  var s3_access_key string
  var s3_secret_key string
  var s3_bucket string
  
  var gcs_creds string
  var gcs_bucket string

  var log_path string
  var log_level string

  config_cmd := &cobra.Command {
    Use: "set",
    Short: "Set Config default variables for hyperflow",
    Long: `Set Config default variables for hyperflow`,
    Run: func(cmd *cobra.Command, args []string) {
       

      var c *config.Config
      c, err := config.NewConfig("", "", config_path)

      if err != nil {
        exitWithError(err.Error())
      }

      if storage_opt != "" {
        c.ObjStorage.StorageTarget = config.StorageTarget(storage_opt)
      }
      
      if log_path != "" {
        c.LogPath = log_path
      }

      if log_level != "" {
        c.LogLevel = log_level
      }
      _, err = config.SaveConfig(c, config_path)
      if err != nil {
        exitWithError(err.Error())
      }
    },
  }

  // variable sets HFLOW_CONFIG_PATH variable
  config_cmd.Flags().StringVarP(&config_path, "config", "c", "", "config path") 

  config_cmd.Flags().StringVar(&storage_opt, "storage-option", "", "Storage backend. e.g. GCS or S3")  
  config_cmd.Flags().StringVar(&log_path, "log-path", "", "log path")
  config_cmd.Flags().StringVar(&log_level, "log-level", "5", "log level")

  s3_cmd := &cobra.Command {
    Use: "s3",
    Short: "Set S3 Config default variables for hyperflow",
    Long: `Set S3 Config default variables for hyperflow`,
    Run: func(cmd *cobra.Command, args []string) {
      
      var c *config.Config
      c, err := config.NewConfig("", "", config_path)
      if err != nil {
        exitWithError(err.Error())
      }
      
      if c.ObjStorage.S3 == nil {
        c.ObjStorage.S3 = &config.S3Config {}
      }

      if s3_creds != "" {
        c.ObjStorage.S3.CredPath = s3_creds 
      }

      if s3_access_key != "" {
        c.ObjStorage.S3.AccessKey = s3_access_key
      }

      if s3_secret_key != "" {
        c.ObjStorage.S3.SecretKey = s3_secret_key
      }

      if s3_bucket != "" {
        c.ObjStorage.S3.Bucket = s3_bucket
      }

      _, err = config.SaveConfig(c, config_path)
      if err != nil {
        exitWithError(err.Error())
      }
    },
  }  
  s3_cmd.Flags().StringVarP(&config_path, "config", "c", "", "config path") 

  s3_cmd.Flags().StringVarP(&s3_creds, "file", "f", "", "S3 storage credentials Path")
  s3_cmd.Flags().StringVarP(&s3_access_key, "access-key", "", "", "S3 storage Access key")
  s3_cmd.Flags().StringVarP(&s3_secret_key, "secret-key", "", "", "S3 storage Secret key")
  s3_cmd.Flags().StringVarP(&s3_bucket, "bucket", "", "", "S3 storage Bucket")

  config_cmd.AddCommand(s3_cmd)

  gcs_cmd := &cobra.Command {
    Use: "gcs",
    Short: "Set GCS Config default variables for hyperflow",
    Long: `Set GCS Config default variables for hyperflow`,
    Run: func(cmd *cobra.Command, args []string) {
      
      var c *config.Config
      c, err := config.NewConfig("", "", config_path)
      if err != nil {
        exitWithError(err.Error())
      }
      
      if c.ObjStorage.Gcs == nil {
        c.ObjStorage.Gcs = &config.GcsConfig {}
      }

      if gcs_creds != "" {
        c.ObjStorage.Gcs.CredsPath = gcs_creds
      }

      if gcs_bucket != "" {
        c.ObjStorage.Gcs.Bucket = gcs_bucket
      }

      _, err = config.SaveConfig(c, config_path)
      if err != nil {
        exitWithError(err.Error())
      }
    },
  }   

  gcs_cmd.Flags().StringVarP(&config_path, "config", "c", "", "config path") 
  gcs_cmd.Flags().StringVarP(&gcs_creds, "file", "f", "", "GCS storage credentials")
  gcs_cmd.Flags().StringVarP(&gcs_bucket, "bucket", "", "", "GCS storage Bucket")

  config_cmd.AddCommand(gcs_cmd)

  var db_name string
  var db_driver string
  var db_user string
  var db_pass string 
  var db_file string 

  db_cmd := &cobra.Command {
    Use: "db",
    Short: "Set DB Config default variables for hyperflow",
    Long: `Set DB Config default variables for hyperflow`,
    Run: func(cmd *cobra.Command, args []string) {
      var c *config.Config
      c, err := config.NewConfig("", "", config_path)
      if err != nil {
        exitWithError(err.Error())
      }
      
      if c.DB == nil {
        c.DB = &config.DBConfig {}
      }

      if db_driver != "" {
        c.DB.Driver = config.DbTarget(db_driver)
      }

      if db_name != "" {
        c.DB.Name = db_name
      }

      if db_user != "" {
        c.DB.User = db_user
      }

      if db_pass != "" {
        c.DB.Pass = db_pass
      }

      if db_file != "" {
        c.DB.DataDirPath = db_file
      }

      _, err = config.SaveConfig(c, config_path)
      if err != nil {
        exitWithError(err.Error())
      }
    }, 
  }

  db_cmd.Flags().StringVarP(&config_path, "config", "c", "", "config path") 

  db_cmd.Flags().StringVarP(&db_driver, "driver", "d", "POSTGRES", "Type of database. eg. POSTGRES")
  db_cmd.Flags().StringVarP(&db_name, "name", "n", "", "Name of database")
  db_cmd.Flags().StringVarP(&db_user, "user", "u", "", "Username of database")
  db_cmd.Flags().StringVarP(&db_pass, "pass", "p", "", "Name of database")
  db_cmd.Flags().StringVarP(&db_file, "file", "f", "", "file that contains connection string")

  config_cmd.AddCommand(db_cmd)

  var kube_config_path string
  var kube_config_url string

  kube_cmd := &cobra.Command {
    Use: "kube",
    Short: "Set Kubernetes Config default variables for hyperflow",
    Long: `Set Kubernetes Config default variables for hyperflow`,
    Run: func(cmd *cobra.Command, args []string) {
      
      var c *config.Config
      c, err := config.NewConfig("", "", config_path)
      if err != nil {
        exitWithError(err.Error())
      }
      
      if c.K8 == nil {
        c.K8 = &config.KubeConfig {}
      }

      if kube_config_path != "" {
        c.K8.Path = kube_config_path
      }
      
      _, err = config.SaveConfig(c, config_path)
      if err != nil {
        exitWithError(err.Error())
      }
    },
  }   
  kube_cmd.Flags().StringVarP(&config_path, "config", "c", "", "config path") 

  kube_cmd.Flags().StringVarP(&kube_config_path, "path", "", "", "Kubernetes config file path")
  kube_cmd.Flags().StringVarP(&kube_config_url, "url", "", "", "Kubernetes config file url")

  config_cmd.AddCommand(kube_cmd)

  return config_cmd
}


func taskCmds() ([]*cobra.Command) {
  var tasks []*cobra.Command

  tasks = []*cobra.Command{}

  return tasks

}

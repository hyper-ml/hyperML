package main



import(
  "os"
  "github.com/spf13/cobra"

  "hyperview.in/server/base"
  "hyperview.in/server/rest"
  "hyperview.in/server/config"
)


func RootCmd() (*cobra.Command, error) {
  var addr string
  var config string 

  root_cmd := &cobra.Command{
    Use: os.Args[0],
    Short: "hyperflow Server Command line",
    Long: `hyperflow Server Command line`,
    Run: func(cmd *cobra.Command, args []string) {
      base.Info("Starting Server... ")
      rest.StartServer(addr)      
    },
  }

  root_cmd.Flags().StringVarP(&addr, "listen", "l", "", "Listen Address with port")
  root_cmd.Flags().StringVarP(&config, "config", "c", "", "Path to config file. Defaults to $HOME/.hflow")  

  root_cmd.AddCommand(configCmd())

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
      
      if config_path != "" {
        config.SetConfigPath(config_path)
        exitWithSuccess()
      }

      var c *config.Config
      c, err := config.GetConfig()

      if err != nil {
        exitWithError(err.Error())
      }

      
      if storage_opt != "" {
        c.StorageOption = storage_opt
      }
      
      if log_path != "" {
        c.LogPath = log_path
      }

      if log_level != "" {
        c.LogLevel = log_level
      }
      err = config.UpdateConfig(c)
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
      c, err := config.GetConfig()
      if err != nil {
        exitWithError(err.Error())
      }
      
      if c.S3 == nil {
        c.S3 = &config.S3Config {}
      }

      if s3_creds != "" {
        c.S3.CredPath = s3_creds 
      }

      if s3_access_key != "" {
        c.S3.AccessKey = s3_access_key
      }

      if s3_secret_key != "" {
        c.S3.SecretKey = s3_secret_key
      }

      if s3_bucket != "" {
        c.S3.Bucket = s3_bucket
      }

      err = config.UpdateConfig(c)
      if err != nil {
        exitWithError(err.Error())
      }
    },
  }  

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
      c, err := config.GetConfig()
      if err != nil {
        exitWithError(err.Error())
      }
      
      if c.Gcs == nil {
        c.Gcs = &config.GcsConfig {}
      }

      if gcs_creds != "" {
        c.Gcs.CredPath = gcs_creds
      }

      if gcs_bucket != "" {
        c.Gcs.Bucket = gcs_bucket
      }

      err = config.UpdateConfig(c)
      if err != nil {
        exitWithError(err.Error())
      }
    },
  }   

  gcs_cmd.Flags().StringVarP(&gcs_creds, "file", "f", "", "GCS storage credentials")
  gcs_cmd.Flags().StringVarP(&gcs_bucket, "bucket", "", "hyperflow001", "GCS storage Bucket")

  config_cmd.AddCommand(gcs_cmd)

  var db_name string
  var db_type string
  var db_user string
  var db_pass string 
  var db_file string 

  db_cmd := &cobra.Command {
    Use: "db",
    Short: "Set DB Config default variables for hyperflow",
    Long: `Set DB Config default variables for hyperflow`,
    Run: func(cmd *cobra.Command, args []string) {
      var c *config.Config
      c, err := config.GetConfig()
      if err != nil {
        exitWithError(err.Error())
      }
      
      if c.DB == nil {
        c.DB = &config.DbConfig {}
      }

      if db_driver != "" {
        c.DB.Driver = db_driver
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
        c.DB.File = db_file
      }

      err = config.UpdateConfig(c)
      if err != nil {
        exitWithError(err.Error())
      }
    }, 
  }
  db_cmd.Flags().StringVarP(&db_driver, "Driver", "t", "POSTGRES", "Type of database. eg. POSTGRES")
  db_cmd.Flags().StringVarP(&db_name, "name", "n", "", "Name of database")
  db_cmd.Flags().StringVarP(&db_user, "user", "u", "", "Username of database")
  db_cmd.Flags().StringVarP(&db_pass, "pass", "p", "", "Name of database")
  db_cmd.Flags().StringVarP(&db_file, "file", "f", "", "file that contains connection string")

  config_cmd.AddCommand(db_cmd)

  return config_cmd
}



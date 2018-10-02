package main



import(
  "os"
  "github.com/spf13/cobra"

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
      rest.ServerMain(addr)      
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

  config_cmd := &cobra.Command {
    Use: "set",
    Short: "Set Config default variables for hyperflow",
    Long: `Set Config default variables for hyperflow`,
    Run: func(cmd *cobra.Command, args []string) {
      
      if config_path != "" {
        config.SetConfigPath(config_path)
      }

      var c *config.Config
      c, err := config.GetConfig()
      if err != nil {
        exitWithError(err)
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
        exitWithError(err)
      }
    }
  }

  // variable sets HFLOW_CONFIG_PATH variable
  config_cmd.PersistentFlags().StringVarP(&config_path, "config", "c", "", "config path") 

  config_cmd.PersistentFlags().StringVarP(&storage_opt, "storage", "sb", "", "Storage backend. e.g. GCS or S3")  
  config_cmd.PersistentFlags().StringVarP(&log_path, "log-path", "", "", "log path")
  config_cmd.PersistentFlags().StringVarP(&log_level, "log-level", "", "5", "log level")

  s3_cmd := &cobra.Command {
    Use: "s3",
    Short: "Set S3 Config default variables for hyperflow",
    Long: `Set S3 Config default variables for hyperflow`,
    Run: func(cmd *cobra.Command, args []string) {
      
      var c *config.Config
      c, err := config.GetConfig()
      if err != nil {
        exitWithError(err)
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
        exitWithError(err)
      }
    }
  }  

  s3_cmd.Flags().StringVarP(&s3_creds, "cred-path", "", "", "S3 storage credentials Path")
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
        exitWithError(err)
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
        exitWithError(err)
      }
    }
  }   

  gcs_cmd.Flags().StringVarP(&gcs_creds, "credentials", "", "", "GCS storage credentials")
  gcs_cmd.Flags().StringVarP(&gcs_bucket, "bucket", "", "", "GCS storage Bucket")

  config_cmd.AddCommand(gcs_cmd)

  return config_cmd
}



func exitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}  

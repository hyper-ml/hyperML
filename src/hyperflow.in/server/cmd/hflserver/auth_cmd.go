package main

import( 
  "fmt"
  "github.com/spf13/cobra"
  "hyperflow.in/server/pkg/base"
  db_pkg "hyperflow.in/server/pkg/db"
  auth_pkg "hyperflow.in/server/pkg/auth"
  config_pkg "hyperflow.in/server/pkg/config"
)

func checkReqParams(values ...string) error {
  
  for _, value := range values {
    if value == "" {
      return fmt.Errorf("One or more mandatory params are missing")
    }
  }
  return nil
}

func authCommand() (*cobra.Command) {
  var user_name string
  var user_type string
  var txtpassword string
  var email string
  var path string

  auth_cmd := &cobra.Command {
    Use: "create_user",
    Short: "Create a new user ",
    Long: `Create a new user `,
    Run: func(cmd *cobra.Command, args []string) {
      // create db context 
      // create auth server
      // create user 
      if err := checkReqParams(user_name, txtpassword); err != nil {
        cmd.Usage()
        exitWithError(err.Error())
      }

      cfg, err := config_pkg.NewConfig("", "", path)
      if err != nil {
        base.Out("[authCommand] failed to get config: ", err)
        exitWithError(err.Error())
      }

      dbc, err := db_pkg.NewDatabaseContext(cfg.DB) 
      if err != nil {
        base.Out("[authCommand] failed to create db context: ", err)
        exitWithError(err.Error())
      }

      auth_server := auth_pkg.NewAuthServer(dbc)
      _, err = auth_server.CreateTypedUser(user_type, user_name, email, txtpassword)  
      if err != nil {
        base.Out("[authCommand] failed to create user: ", err)
        exitWithError(err.Error())
      }  
    },
  }

  auth_cmd.Flags().StringVarP(&path, "config-path", "", "", "") 
  
  auth_cmd.Flags().StringVarP(&user_type, "type", "t", "Standard", "") 
  auth_cmd.Flags().StringVarP(&user_name, "name", "n", "", "") 
  auth_cmd.Flags().StringVarP(&txtpassword, "pass", "p", "", "") 
  auth_cmd.Flags().StringVarP(&email, "email", "e", "", "")

  return auth_cmd
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	authpkg "github.com/hyper-ml/hyperml/server/pkg/auth"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	configpkg "github.com/hyper-ml/hyperml/server/pkg/config"
	dbpkg "github.com/hyper-ml/hyperml/server/pkg/db"
	"github.com/hyper-ml/hyperml/server/pkg/qs"
	"github.com/hyper-ml/hyperml/server/pkg/types"
)

func checkReqParams(values ...string) error {

	for _, value := range values {
		if value == "" {
			return fmt.Errorf("One or more mandatory params are missing")
		}
	}
	return nil
}

func authCommand() *cobra.Command {
	var userName string
	var userType string
	var txtpassword string
	var email string
	var path string

	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User related functions",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	createuser := &cobra.Command{
		Use:   "create",
		Short: "Create a new admin user ",
		Long:  `Create a new admin user `,
		Run: func(cmd *cobra.Command, args []string) {

			if err := checkReqParams(userName, txtpassword); err != nil {
				cmd.Usage()
				exitWithError(err.Error())
			}

			cfg, err := configpkg.NewConfig("", 0, path)
			if err != nil {
				base.Out("[authCommand] failed to get config: ", err)
				exitWithError(err.Error())
			}

			dbc, err := dbpkg.NewDatabaseContext(cfg.DB)
			if err != nil {
				base.Out("[authCommand] failed to create db context: ", err)
				exitWithError(err.Error())
			}

			qs := qs.NewQueryServer(dbc)
			if err != nil {
				base.Out("[authCommand] failed to create common query context: ", err)
				exitWithError(err.Error())
			}

			authServer := authpkg.NewAuthServer(qs, cfg.NoAuth)
			user, err := authServer.CreateTypedUser(types.UserType(userType), userName, email, txtpassword)
			if err != nil {
				base.Out("[authCommand] failed to create user: ", err)
				exitWithError(err.Error())
			}
			userj, _ := json.Marshal(user)
			base.Out(string(userj))
		},
	}

	createuser.Flags().StringVarP(&path, "config-path", "", "", "")

	createuser.Flags().StringVarP(&userType, "type", "", "Admin", "")
	createuser.Flags().StringVarP(&userName, "name", "", "", "")
	createuser.Flags().StringVarP(&txtpassword, "pass", "", "", "")
	createuser.Flags().StringVarP(&email, "email", "", "", "")

	userCmd.AddCommand(createuser)

	return userCmd
}

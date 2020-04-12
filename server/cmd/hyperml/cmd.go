package main

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/seed"

	configpkg "github.com/hyper-ml/hyperml/server/pkg/config"
	"github.com/hyper-ml/hyperml/server/pkg/rest"
)

// RootCmd : Root Command Handler
func RootCmd() (*cobra.Command, error) {
	var ip string
	var port int
	var configpath string
	var ssl bool

	rootcmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Start/Stop hyperML Server ",
		Long:  `Start/Stop hyperML Server `,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	servercmd := &cobra.Command{
		Use:   "server",
		Short: "Start/Stop hyperML Server ",
		Long:  `Start/Stop hyperML Server `,
		Run: func(cmd *cobra.Command, args []string) {
			base.Info("Starting Server... ")
			rest.StartServer(ip, port, ssl, configpath)
		},
	}

	servercmd.PersistentFlags().StringVarP(&ip, "listen-ip", "", "", "Listen IP with port. Defaults to localhost")
	servercmd.PersistentFlags().IntVarP(&port, "listen-port", "", 3000, "Listen Port. Defaults to 8888.")
	servercmd.PersistentFlags().BoolVarP(&ssl, "ssl", "", false, "SSL true or false")
	servercmd.PersistentFlags().StringVarP(&configpath, "config", "c", "", "Path to config file. Defaults to $HOME/.hflow")
	rootcmd.AddCommand(servercmd)

	initcmd := &cobra.Command{
		Use:   "init",
		Short: "Initializes Seed Data (e.e. admin User). Please run only after fresh install.",
		Run: func(cmd *cobra.Command, args []string) {
			conf, err := configpkg.NewConfig("", 0, configpath)
			if err != nil {
				base.Out("[init] failed to find config: ", err)
				exitWithError(err.Error())
			}
			base.Out("Found config file. Seeding data now...")
			if err := seed.Do(conf); err != nil {
				exitWithError(err.Error())
			}
		},
	}
	initcmd.PersistentFlags().StringVarP(&configpath, "config", "c", "", "Path to config file. Defaults to $HOME/.hflow")

	rootcmd.AddCommand(initcmd)
	rootcmd.AddCommand(authCommand())
	return rootcmd, nil
}

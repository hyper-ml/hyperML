package main

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/hyper-ml/hyperml/server/pkg/base"
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
			base.Info("Starting Server... ")

			rest.StartServer(ip, port, ssl, configpath)
		},
	}

	rootcmd.PersistentFlags().StringVarP(&ip, "listen-ip", "", "", "Listen IP with port. Defaults to localhost")
	rootcmd.PersistentFlags().IntVarP(&port, "listen-port", "", 3000, "Listen Port. Defaults to 8888.")
	rootcmd.PersistentFlags().BoolVarP(&ssl, "ssl", "", false, "SSL true or false")
	rootcmd.PersistentFlags().StringVarP(&configpath, "config", "c", "", "Path to config file. Defaults to $HOME/.hflow")

	rootcmd.AddCommand(authCommand())

	return rootcmd, nil
}

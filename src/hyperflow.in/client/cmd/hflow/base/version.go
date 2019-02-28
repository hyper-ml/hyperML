package base


import (
  "fmt"
  "github.com/spf13/cobra"
)

 

var version_cmd = &cobra.Command{
  Use:   "version",
  Short: "Print the version number of Hyperflow",
  Long:  `All software has versions. This is Hyperflow's`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Hyperflow client v0.1")
  },
}
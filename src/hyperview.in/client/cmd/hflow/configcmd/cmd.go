package configcmd


import (
  "fmt"
  "os"
  "strings"
  "path/filepath"
  "github.com/spf13/cobra"

  "hyperview.in/client/config"
)

func exitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}

func ConfigCommand() *cobra.Command {
  var api_server string

  config_cmd := &cobra.Command{
    Use: "config",
    Short: "config parameters ",
    Long: `config parameters `, 
    Run: func(cmd *cobra.Command, args []string) {
      cmd.Usage()
    },
  }

  set_cmd :=  &cobra.Command{
    Use: "set",
    Short: "set config parameters ",
    Long: `set config parameters `, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      err := config.WriteRepoParams(current_dir, "API_SERVER", api_server)
      if err != nil {
        exitWithError(err.Error())
      }
      v, _ := config.ReadRepoParams(current_dir, "API_SERVER")
      fmt.Println("value: ", v)
    },
  }

  set_cmd.Flags().StringVar(&api_server, "api-server", "localhost:8000", "api server")
  config_cmd.AddCommand(set_cmd)

  return config_cmd
}

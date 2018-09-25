package logcmd

import (
  "fmt"
  "os"
  "strings"
  "path/filepath"
  "github.com/spf13/cobra"

  client "hyperview.in/client"
)

func exitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}

func LogCommands() []*cobra.Command {
  var flow_id string

  log_cmd := &cobra.Command{
    Use: "logs",
    Short: "Log for flow Id ",
    //TODO: add command details
    Long: `Log for flow Id `, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      //fmt.Println("directory: ", current_dir)  
      //fmt.Println("task Id: ", flow_id)

      c, _ := client.New(current_dir)
      log_bytes, err := c.RequestLog(flow_id)
      if err != nil {
        exitWithError(err.Error())
      }
      fmt.Println(string(log_bytes))
    },
  }

  log_cmd.Flags().StringVar(&flow_id, "task", "", "task Id")

  var result []*cobra.Command
  result = append(result, log_cmd)
  return result
}



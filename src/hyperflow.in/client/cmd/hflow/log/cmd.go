package log

import (
  "io"
  "fmt"
  "os" 
  "net/url"
  "strings"
  "path/filepath"
  "github.com/spf13/cobra"
  "github.com/gorilla/websocket"

  "hyperflow.in/client/config"
  client "hyperflow.in/client"
  cmd_utils "hyperflow.in/client/cmd/hflow/utils"
)

func isEOFerror(err error) bool{
  err_string := err.Error()
  if strings.Contains(err_string, "EOF") {
    return true
  }
  return false
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

      if flow_id == "" {
        flow_id, _ = config.ReadRepoParams(current_dir, "TASK_ID")
      }
      fmt.Println("Task Id: ", flow_id)
      //todo: add status  
      
      c, _ := client.New(current_dir)
      log_bytes, err := c.RequestLog(flow_id)
      if err != nil {
        cmd_utils.ExitWithError(err.Error())
      }
      fmt.Println(string(log_bytes))
    },
  }

  log_cmd.Flags().StringVar(&flow_id, "task", "", "task Id")
  
  stream_cmd :=  &cobra.Command {
    Use: "stream",
    Short: "Log for flow Id ",
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

      if flow_id == "" {
        flow_id, _ = config.ReadRepoParams(current_dir, "TASK_ID")
      }
      u := url.URL{Scheme: "ws", Host: "localhost:8888", Path: "/ws/log_stream"+ "/" + flow_id}
      c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
      if err != nil {
        fmt.Println("failed to dialup websocket", err)
        return
      }
      defer c.Close()

      for {
        _, message, err := c.ReadMessage()
        if err != nil {
          if err == io.EOF || isEOFerror(err) {
            return
          }
          fmt.Println("Error reading log:", err)
          return
        }
        fmt.Printf(string(message))
      }
    },
  }

  stream_cmd.Flags().StringVar(&flow_id, "task", "", "task Id")

  log_cmd.AddCommand(stream_cmd)

  var result []*cobra.Command
  result = append(result, log_cmd)
  return result
}



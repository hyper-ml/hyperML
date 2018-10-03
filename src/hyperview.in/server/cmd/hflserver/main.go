package main
import(
  "os"
  "fmt"
  "strings"

)


func exitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}  

func exitWithSuccess() {
  os.Exit(0)
}

func main() {
  var err error

  rootCommand, err := RootCmd()
  if err != nil {
    exitWithError(err.Error())
  }

  rootCommand.Execute()
  
}
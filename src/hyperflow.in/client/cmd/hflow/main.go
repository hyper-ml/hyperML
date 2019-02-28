package main



import(
  "hyperflow.in/client/cmd/hflow/base"
	cmd_utils "hyperflow.in/client/cmd/hflow/utils"
)


func main() {
  var err error

  rootCommand, err := base.RootCmd()
  if err != nil {
    cmd_utils.ExitWithError(err.Error())
  }

  rootCommand.Execute()
  
}
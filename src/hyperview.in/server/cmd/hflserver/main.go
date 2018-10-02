package main


func main() {
  var err error

  rootCommand, err := cmd.RootCmd()
  if err != nil {
    cmd.ExitWithError(err.Error())
  }

  rootCommand.Execute()
  
}
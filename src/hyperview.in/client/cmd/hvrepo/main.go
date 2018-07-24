package main



import(
	"hyperview.in/client/cmd/hvrepo/cmd"
)


func main() {
  var err error

  rootCommand, err := cmd.RunnerCmd()
  if err != nil {
    cmd.ExitWithError(err.Error())
  }

  rootCommand.Execute()
  
}
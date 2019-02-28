package base 

import( 
  "os"   
  "fmt" 
  "bufio"
  "syscall"
  "strings"
  "path/filepath"


  "golang.org/x/crypto/ssh/terminal" 
  "github.com/spf13/cobra"

  client "hyperflow.in/client"
  user_config "hyperflow.in/client/config"

  cmd_utils "hyperflow.in/client/cmd/hflow/utils"

)

func getUser() string {
  var user_name string
  var err error
  consolereader := bufio.NewReader(os.Stdin)
  fmt.Print("Please enter hyperflow user name: ")

  user_name, err = consolereader.ReadString('\n') 
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  } 
  return strings.TrimSuffix(user_name, "\n")
}

func getpwd() string {
  
  fmt.Printf("password: ")
  bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
  password := string(bytePassword)
  fmt.Println()
  return password
}


var auth_cmd = &cobra.Command{
  Use: "login",
  Short: "hyperflow login command ",
  Long: `hyperflow login command  `, 
  Run: func(cmd *cobra.Command, args []string) {
    user_name := getUser()
    password := getpwd()
    current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
    c, _ := client.New(current_dir)

    jwt, err := c.Authenticate(user_name, password)
    if err != nil {
      cmd_utils.ExitWithError(err.Error())
    } 
     _ = user_config.SetJwt(jwt) 
  },
}



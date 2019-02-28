package base


import (
	"os"
  "fmt"
)

 

func SetEnvVar(name, value string) error {
  os.Setenv(name, value)
  fmt.Println("Environment value: ", name, os.Getenv(name))

  return nil
}

func GetEnvVar(name string) string {
  return os.Getenv(name)
}

func GetEnv(name string) string{
	return os.Getenv(name)
}
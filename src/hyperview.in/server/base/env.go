package base


import (
	"os"
)



func SetEnv() {
  os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/apple/MyProjects/creds/hw-storage-75d060e8419a.json")
  os.Setenv("GOOGLE_STORAGE_BUCKET", "hyperview001")
}


func SetEnvVar(name, value string) error {
  os.SetEnv(name, value)
  return nil
}

func GetEnvVar(name string) string {
  return os.GetEnv(name)
}

func GetEnv(name string) string{
	return os.Getenv(name)
}
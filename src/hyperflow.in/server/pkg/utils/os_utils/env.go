package os_utils

import(
  "os"
)

func GetOsEnvVar(name string) string {
  return os.Getenv(name)
}
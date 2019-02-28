package fs

import(
  "fmt"
)


func emptyFileMapError() error {
  return fmt.Errorf("File Map is empty")
}

func missingParamsError(plist string) error {
  return fmt.Errorf("One or more required parameter values are missing: %s", plist)
}
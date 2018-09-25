package fs

import(
  "fmt"
)


func emptyFileMapError() error {
  return fmt.Errorf("File Map is empty")
}
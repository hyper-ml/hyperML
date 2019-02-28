package utils

import(
  "strings"
  "hyperflow.in/server/pkg/base"
)


func PrintHeader(title string) {
  base.Println("")
  base.Println(strings.Repeat("*", 50))
  space_after := 50 - len(title) - 4
  if space_after < 0 {
    space_after = 1
  }

  title_string := "*" + strings.Repeat(" ", 2) + title + strings.Repeat(" ", space_after) + "*"
  base.Println(title_string)
  base.Println(strings.Repeat("*", 50))
  base.Println("")
}
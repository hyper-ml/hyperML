package base


import(
  "fmt"
  "runtime"
)


func HomeDir() (string, error) {
  var home string
  if runtime.GOOS == "windows" {
    return "TODO", nil
  } else {
    return unixHome()
  }
  return home, fmt.Errorf("unknown home")
}


// expand ~ in home directory
func unixHome() (string, error) {
  home := GetEnv("HOME")
  if home != "" {
    return home, nil
  }

  // TODO - try os commands 
  return home, fmt.Errorf("Failed to find home directory - HOME env not set")
}


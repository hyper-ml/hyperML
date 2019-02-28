package os_utils
  
import(
  "os"
  "fmt"
  "path/filepath"
  "hyperflow.in/server/pkg/base"
)


func PathExists(p string) bool {
  
  if p == "" {
    return false 
  }

  if _, err := os.Stat(p); os.IsNotExist(err) {
    return false
  }
  return true
}

func K8ConfigPath() string {
  home, _ := base.HomeDir()
  return filepath.Join(home, ".kube", "config")
}

func K8ConfigValidPath() (p string, fnerr error) {
  p = K8ConfigPath()

  if p != "" {
    if PathExists(p) {
      return p, nil
    } else {
      fnerr = fmt.Errorf("No valid kubernetes config found: ", p)
      return 
    }
  }

  return p, fmt.Errorf("Kubernetes config not found in $HOME dir")
}


func HFConfigDefaultPath() string {
  home, _  := base.HomeDir() 
  return filepath.Join(home, ".hflserver")
}

func HFConfigDefaultPathV() string {
  home, _  := base.HomeDir()
  p         := filepath.Join(home, ".hflserver")
  
  if PathExists(p) {
    return p
  } 

  return ""
}


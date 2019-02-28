package base

import ( 
  "fmt"
  //log "github.com/sirupsen/logrus"
)


func Log(format string, args ...interface{}) {
  //log.Infof(format, args)
  //fmt.Println(format, args)
}

func Info(format string, args ...interface{}) {
  //log.Infof(format, args)
  //fmt.Println(format, args)
}

func Warn(format string, args ...interface{}) {
  //log.Warnf(format, args)
  fmt.Println(format, args)

}


func Error(format string, args ...interface{}) {
  //log.Errorf(format, args)
  fmt.Println(format, args)
}


func Debug(format string, args ...interface{}) {
  //log.Debugf(format, args)
  fmt.Println(format, args)
}

func Out(format string, args ...interface{}) {
  fmt.Println(format, args)
}
 
func Println(format string, args ...interface{}) {
  fmt.Println(format, args)
} 


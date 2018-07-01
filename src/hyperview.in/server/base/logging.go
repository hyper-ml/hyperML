package base

import (
  "fmt"
)

type LoggerConfig struct {
	logLevel int
}

type LogKey uint64

const DefaultLogLevel = 0

const (
  infoLevel = iota
  debugLevel
  errorLevel
)

func Log(format string, args ...interface{}) {
  logTo(DefaultLogLevel, 0, format, args)
}


// TODO: log levels and writing to file
func LogInfo(format string, args ...interface{}) {
  logTo(infoLevel, 0, format, args)
}

func logTo(logLevel LogLevel, logKey LogKey, format string, args ...interface{}) {
  fmt.Println(format, args)
}


func Errorf(format string, args ...interface{}) {
  fmt.Errorf(format, args)
}
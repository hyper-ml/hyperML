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
  WarnLevel
  errorLevel
)

func Log(format string, args ...interface{}) {
  logTo(DefaultLogLevel, 0, format, args)
}


func Info(format string, args ...interface{}) {
  logTo(DefaultLogLevel, 2, format, args)
}

func Warn(format string, args ...interface{}) {
  logTo(DefaultLogLevel, 2, format, args)
}


func Error(format string, args ...interface{}) {
  logTo(DefaultLogLevel, 2, format, args)
}


func Debug(format string, args ...interface{}) {
  logTo(DefaultLogLevel, 1, format, args)
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
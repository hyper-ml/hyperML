package base

import (
  "fmt"
)

type LoggerConfig struct {
	logLevel int
}

type LogKey uint64

const LogLevel int = 0
const defaultLevel int = 4

const (
  InfoLevel = iota
  DebugLevel
  WarnLevel
  ErrorLevel
)



func Log(format string, args ...interface{}) {
  logTo(LogLevel, LogKey(defaultLevel), format, args)
}


func Info(format string, args ...interface{}) {
  logTo(LogLevel, InfoLevel, format, args)
}

func Warn(format string, args ...interface{}) {
  logTo(LogLevel, WarnLevel, format, args)
}


func Error(format string, args ...interface{}) {
  logTo(LogLevel, ErrorLevel, format, args)
}


func Debug(format string, args ...interface{}) {
  logTo(LogLevel, DebugLevel, format, args)
}

func Out(format string, args ...interface{}) {
  logTo(LogLevel, InfoLevel, format, args)
}
  

func logTo(logLevel int, logKey LogKey, format string, args ...interface{}) {
  if int(logKey) >= int(logLevel) {
    fmt.Println(format, args)
  }
}


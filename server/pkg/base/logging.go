package base

import (
	"fmt"
	//log "github.com/sirupsen/logrus"
)

// Log : Log messages
func Log(format string, args ...interface{}) {
	//log.Infof(format, args)
	fmt.Println(format, args)
}

// Info : Log Info messages
func Info(format string, args ...interface{}) {
	//log.Infof(format, args)
	fmt.Println(format, args)
}

// Warn : Log warning messages
func Warn(format string, args ...interface{}) {
	//log.Warnf(format, args)
	fmt.Println(format, args)

}

// Error : Log Error messages
func Error(format string, args ...interface{}) {
	//log.Errorf(format, args)
	fmt.Println(format, args)
}

// Debug : Log Debug messages
func Debug(format string, args ...interface{}) {
	//log.Debugf(format, args)
	fmt.Println(format, args)
}

// Out : Log output messages
func Out(format string, args ...interface{}) {
	fmt.Println(format, args)
}

// Println : Print messages
func Println(format string, args ...interface{}) {
	fmt.Println(format, args)
}

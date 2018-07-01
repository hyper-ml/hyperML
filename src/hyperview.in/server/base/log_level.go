package base


import (
	"fmt"
  "sync/atomic"
)

type LogLevel uint32

const (
  LevelZero LogLevel = iota
  LevelError 
  LevelWarn
  LevelInfo
  LevelDebug
  LevelTrace

  levelCount
)


var (
  levelNames = []string{"off", "error", "warn", "info", "debug", "trace"}
  levelShortNames = []string{"OFF", "ERR", "WRN", "INF", "DBG", "TRC"}
)

/* why: use atomic operation so current processes will not be affected by update */
func (l *LogLevel) set(newLevel LogLevel) {
  atomic.StoreUint32((*uint32) (l), uint32(newLevel))
}

/* what: check if log level is enabled */ 
func (l *LogLevel) Enabled(logLevel LogLevel) bool {
  if l == nil {
    return false
  }
  return atomic.LoadUint32((*uint32)(l)) >= uint32(logLevel)
}

func (l LogLevel) String() string {
  if l >= levelCount {
    return fmt.Sprintf("LogLevel(%d)", l)
  }
  return levelNames[l]
}

func (l LogLevel) StringShort() string {
  if l >= levelCount {
    return fmt.Sprintf("LVL(%d)", l)
  }
  return levelShortNames[l]
}


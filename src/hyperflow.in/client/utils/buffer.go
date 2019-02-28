package utils

import (
	"sync"
)

var (
	MaxBufferSize = 20 * 1024 * 1024 ) // 2MB / 1000

var bufferQ = sync.Pool {
	New: func() interface{} {
		return make([]byte, MaxBufferSize/10)
	},
}

func GetBuffer() []byte {
  return bufferQ.Get().([]byte)
}

func PutBuffer(buf []byte) {
  bufferQ.Put(buf)
}


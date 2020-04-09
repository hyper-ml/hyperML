package utils

import (
	"sync"
)

var (
	// MaxMsgSize :
	MaxMsgSize = 20 * 1024 * 1024
) // 2MB / 1000

var bufferQ = sync.Pool{
	New: func() interface{} {
		return make([]byte, MaxMsgSize/10)
	},
}

// GetBuffer :
func GetBuffer() []byte {
	return bufferQ.Get().([]byte)
}

// PutBuffer :
func PutBuffer(buf []byte) {
	bufferQ.Put(buf)
}

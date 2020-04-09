package base

import (
	"strconv"
	"sync"
	"time"
)

// IntMax : Used Locking and tracking waiters
type IntMax struct {
	i  int64
	mu sync.RWMutex
}

func (v *IntMax) String() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return strconv.FormatInt(v.i, 10)
}

// SetIfMax :
func (v *IntMax) SetIfMax(value int64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if value > v.i {
		v.i = value
	}
}

// IntMeanVar :
type IntMeanVar struct {
	count int64 // number of values seen
	mean  int64 // average value
	mu    sync.RWMutex
}

func (v *IntMeanVar) String() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return strconv.FormatInt(v.mean, 10)
}

// AddValue :  Calculates new mean as iterative mean (avoids int overflow)
func (v *IntMeanVar) AddValue(value int64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.count++
	v.mean = v.mean + int64((value-v.mean)/v.count)
}

// AddSince :
func (v *IntMeanVar) AddSince(start time.Time) {
	v.AddValue(time.Since(start).Nanoseconds())
}

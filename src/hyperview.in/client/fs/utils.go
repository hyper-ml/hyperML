package fs

import (
  "io"
  "regexp"

)

func DirNameForRepo(repoName string) string {
  reg, err := regexp.Compile("[^a-zA-Z0-9]+")
  if err != nil {
    return 
  }
  return reg.ReplaceAllString(repoName, "")
}

// customer writer to keep track of writes
type CountWriter struct {
  w io.Writer
  size int64
}

func (c *CountWriter) Write(b []byte) (int, error) {
  n, err := c.w.Write(b)
  c.size += int64(n)
  return n, err
}

// control parallel ops  
type SyncOpLimiter interface {
  Ask()
  Release()
  Wait()
}

type syncOpLimiter struct {
  rooms chan struct {}
}

func (s *syncOpLimiter) Ask() {
  s.rooms <- struct{}{}
}

func (s *syncOpLimiter) Release() {
  <- s.rooms
}

func (s *syncOpLimiter) Wait() {
  for i :=0; i < cap(s.rooms); i++ {
    s.rooms <- struct{}{}
  }
}

type noOpLimiter struct {}

func (n *noOpLimiter) Ask() {}
func (n *noOpLimiter) Release() {}
func (n *noOpLimiter) Wait() {}

func NewOpLimiter(ops int) SyncOpLimiter {
  if ops == 0 {
    return &noOpLimiter{}
  }
  rooms := make(chan struct{}, ops)

  return &syncOpLimiter{rooms}
}
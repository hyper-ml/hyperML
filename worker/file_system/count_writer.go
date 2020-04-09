package file_system

import (
  "io"
)

type CountWriter struct {
  w io.Writer
  size int64
}

func (c *CountWriter) Write(b []byte) (int, error) {
  n, err := c.w.Write(b)
  c.size += int64(n)
  return n, err
}
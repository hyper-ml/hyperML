package code_sync


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
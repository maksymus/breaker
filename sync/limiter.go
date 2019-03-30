package sync

import "sync"

type Limiter interface {
  Take() <-chan *struct{}
  TakeOrNil() *struct{}
  Return(ticket *struct{})
  Size() int
}

// ticket based pool to limit number of invocations
type limiter struct {
  tickets chan *struct{}
  maxSize int
  mutex   sync.RWMutex
}

func NewLimiter(size int) Limiter {
  pool := &limiter{
    make(chan *struct{}, size),
    size,
    sync.RWMutex{},
  }

  for i := 0; i < size; i++ {
    pool.tickets <- &struct{}{}
  }

  return pool
}

func (limiter *limiter) Take() <-chan *struct{} {
  return limiter.tickets
}

func (limiter *limiter) TakeOrNil() *struct{} {
  select {
  case ticket := <-limiter.tickets:
    return ticket
  default:
    return nil
  }
}

// return ticket back to pool
func (limiter *limiter) Return(ticket *struct{}) {
  if ticket == nil {
    return
  }

  limiter.mutex.Lock()
  defer limiter.mutex.Unlock()

  if len(limiter.tickets) < limiter.maxSize {
    limiter.tickets <- ticket
  }
}

// number of free tickets in pool
func (limiter *limiter) Size() int {
  limiter.mutex.RLock()
  defer limiter.mutex.RUnlock()

  return len(limiter.tickets)
}

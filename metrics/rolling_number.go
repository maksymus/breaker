package metrics

import (
  "sync"
  "time"
  "fmt"
)

type Number interface {
  Increment()
  Add(value int64)
  GetValue() int64
  Sum(time.Time) int64
}

// number holds rolling data for time period
// data is stored in circular array of buckets
// bucket stores some metrics - for example number of circuit executions
type rollingNumber struct {
  buckets circularArray
}

// circular array to store buckets
type circularArray struct {
  array    []*bucket
  slots    uint
  period   time.Duration
  mutex    sync.RWMutex
  position int
}

// bucket holding value
type bucket struct {
  value int64
  start time.Time
  adder chan int64
}

// create number with slots each holding data for some period
// for example 10 slots 1 sec period each - data for 10 seconds
func CreateNumber(slots uint, period time.Duration) Number {
  return &rollingNumber{
    buckets: createCircularArray(slots, period),
  }
}

func (number *rollingNumber) Increment() {
  currentBucket := number.buckets.getCurrentBucket()
  currentBucket.adder <- 1
}

func (number *rollingNumber) Add(value int64) {
  currentBucket := number.buckets.getCurrentBucket()
  currentBucket.adder <- value
}

func (number *rollingNumber) GetValue() int64 {
  currentBucket := number.buckets.getCurrentBucket()
  return currentBucket.value
}

func (number *rollingNumber) Sum(time time.Time) int64 {
  return number.buckets.Sum(time)
}

// create circular arrays with
// - size - number of buckets
// - timespan - timespan in ms of data stored in buckets (for example 100ms stored in bucket, 10 buckets - 1 sec array)
func createCircularArray(slots uint, period time.Duration) circularArray {
  return circularArray{
    array:  make([]*bucket, slots),
    slots:  slots,
    period: period,
    mutex:  sync.RWMutex{},
  }
}

func createBucket(time time.Time) *bucket {
  bucket := bucket{0, time, make(chan int64)}

  go func() {
    for {
      select {
      case add := <-bucket.adder:
        bucket.value += add
      }
    }
  }()

  return &bucket
}

func (ca *circularArray) getCurrentBucket() *bucket {
  now := time.Now()

  ca.mutex.Lock()
  defer ca.mutex.Unlock()

  if b := ca.array[ca.position]; b != nil {
    if b.start.UnixNano()+ca.period.Nanoseconds() < now.UnixNano() {
      ca.position = (ca.position + 1) % len(ca.array)
      ca.array[ca.position] = createBucket(now)
    }
  } else {
    ca.array[ca.position] = createBucket(now)
    ca.position = ca.position % len(ca.array)
  }

  return ca.array[ca.position]
}

// returns sum of bucket values for time window of period * slots
func (ca *circularArray) Sum(now time.Time) int64 {
  ca.mutex.Lock()
  defer ca.mutex.Unlock()

  var sum int64

  for _, bucket := range ca.array {
    if bucket != nil {
      // sum of periods in each slot
      numberRollingTime := ca.period.Nanoseconds() * int64(ca.slots)

      // how much time passed since bucket was created
      bucketStartTimePassed := now.Sub(bucket.start).Nanoseconds()

      if bucketStartTimePassed <= numberRollingTime && bucket.start.Before(now) {
        sum += bucket.value
      }
    }
  }

  return sum
}

func main() {
  ca := createCircularArray(1, time.Second)

  bucket1 := ca.getCurrentBucket()
  fmt.Println(bucket1.start)
  bucket2 := ca.getCurrentBucket()
  fmt.Println(bucket2.start)
  bucket3 := ca.getCurrentBucket()
  fmt.Println(bucket3.start)

  time.Sleep(1000 * time.Millisecond)

  bucket4 := ca.getCurrentBucket()
  fmt.Println(bucket4.start)
  bucket5 := ca.getCurrentBucket()
  fmt.Println(bucket5.start)

  time.Sleep(1000 * time.Millisecond)

  bucket6 := ca.getCurrentBucket()
  fmt.Println(bucket6.start)
  bucket7 := ca.getCurrentBucket()
  fmt.Println(bucket7.start)
}

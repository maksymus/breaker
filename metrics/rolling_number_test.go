package metrics

import (
  "testing"
  . "github.com/smartystreets/goconvey/convey"
  "time"
  "sync"
)

func Test_Increment(t *testing.T) {
  Convey("run Increment command", t, func() {
    number := CreateNumber(10, time.Millisecond*100)

    number.Increment()

    Convey("value should be 1", func() {
      So(number.GetValue(), ShouldEqual, 1)
    })
  })
}

func Test_Increment_Multi(t *testing.T) {
  Convey("run Increment command with concurrent calls", t, func() {
    number := CreateNumber(10, time.Minute)

    expected := 100

    wg := sync.WaitGroup{}
    wg.Add(expected)

    for i := 0; i < expected; i++ {
      go func() {
        defer wg.Done()
        number.Increment()
      }()
    }

    wg.Wait()
    time.Sleep(time.Second)

    Convey("value should match number of increments", func() {
      So(number.GetValue(), ShouldEqual, expected)
    })
  })
}

func Test_Sum(t *testing.T) {
  Convey("run Sum command", t, func() {
    number := CreateNumber(10, time.Millisecond*100)

    number.Increment()
    time.Sleep(time.Millisecond * 10)
    value1 := number.GetValue()

    time.Sleep(time.Millisecond * 100)

    number.Increment()
    time.Sleep(time.Millisecond * 10)
    value2 := number.GetValue()

    Convey("value1 should be correct", func() {
      So(value1, ShouldEqual, 1)
    })

    Convey("value2 should be correct", func() {
      So(value2, ShouldEqual, 1)
    })

    Convey("sum over period of 1 sec", func() {
      So(number.Sum(time.Now()), ShouldEqual, 2)
    })
  })
}

func Test_Sum_Expired(t *testing.T) {
  Convey("run Sum command ", t, func() {
    number := CreateNumber(10, time.Millisecond*10)

    number.Increment()
    number.Increment()
    number.Increment()
    time.Sleep(time.Millisecond * 1000)

    Convey("measurement time (10ms * 10 slots = 100ms) expires", func() {
      So(number.Sum(time.Now()), ShouldEqual, 0)
    })
  })
}

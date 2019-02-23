package breaker

import (
  "testing"
  . "github.com/smartystreets/goconvey/convey"
  "time"
  "sync"
  "breaker/metrics"
  "fmt"
)

func Test_allowSingleTest_Sequence(t *testing.T) {
  Convey("circuit singleTest in sequence", t, func() {
    circuit := getCircuit("test")

    allow1 := circuit.allowSingleTest()
    lastTested1 := circuit.lastTested

    allow2 := circuit.allowSingleTest()
    lastTested2 := circuit.lastTested

    // default circuit sleep duration set to 1 sec
    time.Sleep(time.Second)
    allow3 := circuit.allowSingleTest()
    lastTested3 := circuit.lastTested

    Convey("circuit should wait for sleep duration before allowing test", func() {
      So(allow1, ShouldEqual, true)
      So(allow2, ShouldEqual, false)
      So(allow3, ShouldEqual, true)

      So(lastTested1, ShouldEqual, lastTested2)
      So(lastTested2, ShouldBeLessThan, lastTested3)
    })
  })
}

func Test_allowSingleTest_TwoParallelCalls(t *testing.T) {
  Convey("circuit singleTest in two parallel calls", t, func() {
    circuit := getCircuit("test1")

    barrier := sync.WaitGroup{}
    barrier.Add(1)

    testsDone := sync.WaitGroup{}
    testsDone.Add(2)

    var allowTest1 bool
    var allowTest2 bool

    go func() {
      barrier.Wait()
      allowTest1 = circuit.allowSingleTest()
      testsDone.Done()
    }()

    go func() {
      barrier.Wait()
      allowTest2 = circuit.allowSingleTest()
      testsDone.Done()
    }()

    barrier.Done()
    testsDone.Wait()

    Convey("only one test is allowed", func() {
      So(allowTest1 || allowTest2, ShouldEqual, true)
      So(allowTest1 && allowTest2, ShouldEqual, false)
    })
  })
}

func Test_isBroken(t *testing.T) {
  Convey("circuit isBroken test", t, func() {
    type TestCase struct {
      requests int64
      errors   int64
      broken   bool
    }

    cases := []TestCase{
      {10, 0, false},
      {0, 10, true},
      {10, 1, false},
      {10, 9, true},
      {10, 8, true},
    }

    circuit := getCircuit("broken test1")
    ConfigureCircuit(circuit.name, Settings{
      ErrorThreshold: 0.8,
    })

    for _, tc := range cases {
      circuit.metrics = mockMetricsCollector{
        tc.requests,
        tc.errors,
      }

      broken := circuit.isBroken()

      Convey(fmt.Sprintf("circuit with %d requests and %d errors is broken? %t",
        tc.requests, tc.errors, tc.broken), func() {
        So(broken, ShouldEqual, tc.broken)
      })
    }
  })
}

// mocks
type mockMetricsCollector struct {
  requestSum int64
  errorSum   int64
}

func (mock mockMetricsCollector) Requests() metrics.Number {
  return mockNumber{mock.requestSum}
}

func (mock mockMetricsCollector) Errors() metrics.Number {
  return mockNumber{mock.errorSum}
}

func (mock mockMetricsCollector) Rejects() metrics.Number {
  panic("implement me")
}

func (mock mockMetricsCollector) Timeouts() metrics.Number {
  panic("implement me")
}

func (mock mockMetricsCollector) Cancelled() metrics.Number {
  panic("implement me")
}

func (mock mockMetricsCollector) FallbackSuccess() metrics.Number {
  panic("implement me")
}

func (mock mockMetricsCollector) FallbackFailure() metrics.Number {
  panic("implement me")
}

func (mockMetricsCollector) Reset() {
  fmt.Println("reseting")
}

type mockNumber struct {
  sum int64
}

func (mockNumber) Increment() {
  panic("implement me")
}

func (mockNumber) Add(value int64) {
  panic("implement me")
}

func (mockNumber) GetValue() int64 {
  panic("implement me")
}

func (mock mockNumber) Sum(time.Time) int64 {
  return mock.sum
}

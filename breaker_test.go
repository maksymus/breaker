package breaker

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"context"
	"time"
	"breaker/errors"
	"fmt"
	"sync"
)

func Test_Go(t *testing.T) {
	Convey("run Go command ", t, func() {
		resultChan := make(chan interface{}, 1)


		executeCmd := func(ctx context.Context) error {
			resultChan <- 1
			return nil
		}

		errChan := Go("Test_Go", context.Background(), executeCmd, nil)

		Convey("metrics are recorded", func() {
			circuit := getCircuit("Test_Go")
			circuit.mutex.RLock()
			defer circuit.mutex.RUnlock()

			So(circuit.metrics.Requests().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Errors().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackSuccess().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackFailure().Sum(time.Now()), ShouldEqual, 0)
		})

		Convey("reading from that channel should provide the expected value", func() {
			So(<-resultChan, ShouldEqual, 1)
		})

		Convey("no errors should be returned", func() {
			So(len(errChan), ShouldEqual, 0)
		})
	})
}

func Test_Go_Timeout(t *testing.T) {
	Convey("run Go command", t, func() {
		resultChan := make(chan interface{}, 1)

		executeCmd := func(ctx context.Context) error {
			time.Sleep(2 * time.Second)
			return nil
		}

		errChan := Go("Test_Go_Timeout", context.Background(), executeCmd, nil)

		Convey("metrics are recorded", func() {
			circuit := getCircuit("Test_Go_Timeout")
			circuit.mutex.RLock()
			defer circuit.mutex.RUnlock()

			So(circuit.metrics.Requests().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Errors().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Rejects().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Timeouts().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Cancelled().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackSuccess().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackFailure().Sum(time.Now()), ShouldEqual, 0)
		})

		Convey("reading from that channel should provide the expected value", func() {
			So(len(resultChan), ShouldEqual, 0)
		})

		Convey("timeout error", func() {
			So(<-errChan, ShouldResemble, errors.TimeoutError)
		})
	})
}

func Test_Go_Failed(t *testing.T) {
	Convey("run Go command and fail with no failover func defined", t, func() {

		executeCmd := func(ctx context.Context) error {
			return fmt.Errorf("exec failure")
		}

		errChan := Go("Test_Go_Failed", context.Background(), executeCmd, nil)

		Convey("metrics are recorded", func() {
			circuit := getCircuit("Test_Go_Failed")
			circuit.mutex.RLock()
			defer circuit.mutex.RUnlock()

			So(circuit.metrics.Requests().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Errors().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Rejects().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Timeouts().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Cancelled().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackSuccess().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackFailure().Sum(time.Now()), ShouldEqual, 0)
		})

		Convey("error should be returned", func() {
			So(len(errChan), ShouldEqual, 1)
			So(<-errChan, ShouldResemble, fmt.Errorf("exec failure"))
		})
	})
}

func Test_Go_FailoverSuccess(t *testing.T) {
	Convey("run Go command ", t, func() {

		failoverChan := make(chan interface{}, 1)

		executeCmd := func(ctx context.Context) error {
			return fmt.Errorf("exec failure")
		}

		failoverCmd := func(ctx context.Context, err error) error {
			failoverChan <- 1
			return nil
		}

		errChan := Go("Test_Go_FailoverSuccess", context.Background(), executeCmd, failoverCmd)

		Convey("metrics are recorded", func() {
			circuit := getCircuit("Test_Go_FailoverSuccess")
			circuit.mutex.RLock()
			defer circuit.mutex.RUnlock()

			So(circuit.metrics.Requests().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Errors().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Rejects().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Timeouts().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Cancelled().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackSuccess().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.FallbackFailure().Sum(time.Now()), ShouldEqual, 0)
		})

		Convey("failover should be executed", func() {
			So(len(failoverChan), ShouldEqual, 1)
		})

		Convey("no error returned", func() {
			So(len(errChan), ShouldEqual, 0)
		})

	})
}

func Test_Go_FailoverFailure(t *testing.T) {
	Convey("run Go command failover fails", t, func() {

		failoverChan := make(chan interface{}, 1)

		executeCmd := func(ctx context.Context) error {
			return fmt.Errorf("exec failure")
		}

		failoverCmd := func(ctx context.Context, err error) error {
			failoverChan <- 1
			return fmt.Errorf("failover failure")
		}

		errChan := Go("Test_Go_FailoverFailure", context.Background(), executeCmd, failoverCmd)

		Convey("metrics are recorded", func() {
			circuit := getCircuit("Test_Go_FailoverFailure")
			circuit.mutex.RLock()
			defer circuit.mutex.RUnlock()

			So(circuit.metrics.Requests().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Errors().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Rejects().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Timeouts().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Cancelled().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackSuccess().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackFailure().Sum(time.Now()), ShouldEqual, 1)
		})

		Convey("failover should be executed", func() {
			So(len(failoverChan), ShouldEqual, 1)
		})

		Convey("error returned", func() {
			So(len(errChan), ShouldEqual, 1)
			So(<-errChan, ShouldResemble, fmt.Errorf("failover failure"))
		})
	})
}

func Test_Go_MaxConcurrentLimitReached(t *testing.T) {
	Convey("run Go command ", t, func() {
		settings := GetSettings("concurrent")
		settings.MaxConcurrentCalls = 1

		ConfigureCircuit("Test_Go_MaxConcurrentLimitReached", settings)

		// 1st command waits for second to complete
		wg := sync.WaitGroup{}
		wg.Add(1)

		// wait for both commands to complete
		wg1 := sync.WaitGroup{}
		wg1.Add(2)

		resultChan1 := make(chan interface{}, 1)
		resultChan2 := make(chan interface{}, 1)

		executeCmd1 := func(ctx context.Context) error {
			wg.Wait()
			resultChan1 <- 1
			return nil
		}

		executeCmd2 := func(ctx context.Context) error {
			resultChan2 <- 1
			return nil
		}

		var errChan1 chan error
		var errChan2 chan error

		go func() {
			errChan1 = Go("Test_Go_MaxConcurrentLimitReached", context.Background(), executeCmd1, nil)
			wg1.Done()
		}()

		time.Sleep(time.Millisecond * 100)

		go func() {
			errChan2 = Go("Test_Go_MaxConcurrentLimitReached", context.Background(), executeCmd2, nil)
			wg.Done()
			wg1.Done()
		}()

		wg1.Wait()

		Convey("metrics are recorded", func() {
			circuit := getCircuit("Test_Go_MaxConcurrentLimitReached")
			circuit.mutex.RLock()
			defer circuit.mutex.RUnlock()

			So(circuit.metrics.Requests().Sum(time.Now()), ShouldEqual, 2)
			So(circuit.metrics.Errors().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Rejects().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Timeouts().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Cancelled().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackSuccess().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackFailure().Sum(time.Now()), ShouldEqual, 0)
		})

		Convey("first command successful", func() {
			So(len(resultChan1), ShouldEqual, 1)
			So(len(errChan1), ShouldEqual, 0)
		})

		Convey("second command fails with max concurrent error", func() {
			So(len(resultChan2), ShouldEqual, 0)
			So(len(errChan2), ShouldEqual, 1)
			So(<-errChan2, ShouldResemble, errors.ConcurrentLimitError)
		})
	})
}

func Test_Go_ExecPanic(t *testing.T) {
	Convey("run Go command and fail with no failover func defined", t, func() {

		executeCmd := func(ctx context.Context) error {
			panic("invalid data")
		}

		errChan := Go("Test_Go_Panic", context.Background(), executeCmd, nil)

		Convey("metrics are recorded", func() {
			circuit := getCircuit("Test_Go_Panic")
			circuit.mutex.RLock()
			defer circuit.mutex.RUnlock()

			So(circuit.metrics.Requests().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Errors().Sum(time.Now()), ShouldEqual, 1)
			So(circuit.metrics.Rejects().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Timeouts().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.Cancelled().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackSuccess().Sum(time.Now()), ShouldEqual, 0)
			So(circuit.metrics.FallbackFailure().Sum(time.Now()), ShouldEqual, 0)
		})

		Convey("error should be returned", func() {
			So(len(errChan), ShouldEqual, 1)
			So(<-errChan, ShouldResemble, fmt.Errorf("exec panic: invalid data"))
		})
	})
}


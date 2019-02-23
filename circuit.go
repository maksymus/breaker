package breaker

import (
	"sync"
	"breaker/metrics"
	"time"
	"sync/atomic"
	bsync "breaker/sync"
)

var circuits map[string]*circuit
var mutex sync.RWMutex

type eventType string

const (
	success         eventType = "success"
	failure                   = "failure"
	rejected                  = "rejected"
	timeout                   = "timeout"
	cancelled                 = "cancelled"
	fallbackSuccess           = "fallback success"
	fallbackFailure           = "fallback failure"
)

type event struct {
	rootEvent     eventType
	fallbackEvent eventType
}

type circuit struct {
	name       string
	mutex      sync.RWMutex
	metrics    metrics.Collector
	limiter    bsync.Limiter
	lastTested int64 // init to 0
	events     chan event
}

func init() {
	circuits = make(map[string]*circuit)
	mutex = sync.RWMutex{}
}

func getCircuit(name string) *circuit {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := circuits[name]; !ok {
		settings := GetSettings(name)

		circuit := circuit{
			name:    name,
			metrics: metrics.NewCollector(),
			limiter: bsync.NewLimiter(settings.MaxConcurrentCalls),
			events:  make(chan event),
		}

		// listen to events
		go func() {
			// TODO allow to stop and exit loop
			for {
				select {
				case event := <-circuit.events:
					circuit.reportEvent(event)
				}
			}
		}()

		circuits[name] = &circuit
	}

	return circuits[name]
}

func (circuit *circuit) AllowRequest() bool {
	return !circuit.isBroken() || circuit.allowSingleTest()
}

// too many failed requests
func (circuit *circuit) isBroken() bool {
	settings := GetSettings(circuit.name)

	metrics := circuit.metrics

	circuit.mutex.RLock()
	defer circuit.mutex.RUnlock()

	now := time.Now()
	requests := metrics.Requests().Sum(now)
	errors := metrics.Errors().Sum(now)

	if errors == 0 {
		return false
	}

	return float32(errors) / float32(requests) >= settings.ErrorThreshold
}

// try once to check if circuit is restored
func (circuit *circuit) allowSingleTest() bool {
	settings := GetSettings(circuit.name)

	lastTested := atomic.LoadInt64(&circuit.lastTested)
	wakeupTime := lastTested + settings.SleepDuration.Nanoseconds()

	now := time.Now().UnixNano()

	if wakeupTime < now {
		swapped := atomic.CompareAndSwapInt64(&circuit.lastTested, lastTested, now)
		// don't allow if not swapped - other call swapped it
		return swapped
	}

	return false
}

func (circuit *circuit) reportEvent(event event) {
	circuit.mutex.Lock()
	defer circuit.mutex.Unlock()

	metrics := circuit.metrics
	metrics.Requests().Increment()

	if event.rootEvent != success {
		metrics.Errors().Increment()

		switch event.rootEvent {
		case rejected:
			metrics.Rejects().Increment()
		case timeout:
			metrics.Timeouts().Increment()
		case cancelled:
			metrics.Cancelled().Increment()
		}

		switch event.fallbackEvent {
		case fallbackSuccess:
			metrics.FallbackSuccess().Increment()
		case fallbackFailure:
			metrics.FallbackFailure().Increment()
		}
	}
}
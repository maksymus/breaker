package metrics

import "time"

const (
	slots        = 10
	slotDuration = time.Second
)

type Collector interface {
	Reset()
	Requests() Number
	Errors() Number
	Rejects() Number
	Timeouts() Number
	Cancelled() Number
	FallbackSuccess() Number
	FallbackFailure() Number
}

type collector struct {
	requests  Number
	errors    Number

	rejects   Number
	timeouts  Number
	cancelled Number

	fallbackSuccess Number
	fallbackFailure Number
}

func NewCollector() Collector {
	collector := &collector{}
	collector.Reset()
	return collector
}

func (c *collector) Requests() Number {
	return c.requests
}

func (c *collector) Errors() Number {
	return c.errors
}

func (c *collector) Rejects() Number {
	return c.rejects
}

func (c *collector) Timeouts() Number {
	return c.timeouts
}

func (c *collector) Cancelled() Number {
	return c.cancelled
}

func (c *collector) FallbackSuccess() Number {
	return c.fallbackSuccess
}

func (c *collector) FallbackFailure() Number {
	return c.fallbackFailure
}

func (c *collector) Reset() {
	c.requests = CreateNumber(slots, slotDuration)
	c.errors = CreateNumber(slots, slotDuration)

	c.rejects = CreateNumber(slots, slotDuration)
	c.timeouts = CreateNumber(slots, slotDuration)
	c.cancelled = CreateNumber(slots, slotDuration)

	c.fallbackSuccess = CreateNumber(slots, slotDuration)
	c.fallbackFailure = CreateNumber(slots, slotDuration)
}

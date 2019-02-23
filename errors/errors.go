package errors

import "fmt"

var (
	ConcurrentLimitError = fmt.Errorf("concurrent calls limit reached")
	TimeoutError         = fmt.Errorf("timeout")
	CircuitBrokenError   = fmt.Errorf("circuit is broken")
	CancelledError       = fmt.Errorf("cancelled")
)

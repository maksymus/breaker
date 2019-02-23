package breaker

import (
  "time"
  "context"
  "breaker/errors"
  "fmt"
)

func Go(name string, context context.Context,
  exec func(context.Context) error,
  fail func(context.Context, error) error) chan error {
  circuit := getCircuit(name)

  timeout := GetSettings(circuit.name).Timeout
  timer := time.NewTicker(timeout)

  executor := executor{
    execCmd: exec,
    failCmd: fail,
    start:   time.Now(),
    done:    make(chan bool, 1),
    err:     make(chan error, 1),
  }

  ticket := circuit.limiter.TakeOrNil()
  defer circuit.limiter.Return(ticket)

  if ticket == nil {
    executor.fail(context, circuit, errors.ConcurrentLimitError)
  } else {
    go func() {
      executor.execute(context, circuit)
    }()

    select {
    case <-context.Done():
      executor.fail(context, circuit, errors.CancelledError)
    case <-timer.C:
      executor.fail(context, circuit, errors.TimeoutError)
    case <-executor.done:
      break
    }
  }

  return executor.err
}

type executor struct {
  execCmd func(context.Context) error
  failCmd func(context.Context, error) error
  start   time.Time
  end     time.Time
  done    chan bool
  err     chan error
  failed  bool
}

func (e *executor) execute(ctx context.Context, circuit *circuit) {
  defer func() {
    e.done <- true
    close(e.done)

    if !e.failed {
      e.end = time.Now()
      circuit.reportEvent(event{rootEvent: success})
    }
  }()

  err := e.execCmdWrapper(ctx, circuit)
  if err != nil {
    e.fail(ctx, circuit, err)
  }
}

func (e *executor) fail(ctx context.Context, circuit *circuit, execError error) {
  e.failed = true

  var failError error
  defer func() {
    e.end = time.Now()

    if e.failCmd != nil {
      circuit.reportEvent(event{translateError(execError), translateFallbackError(failError)})
    } else {
      circuit.reportEvent(event{rootEvent: translateError(execError)})
    }

    close(e.err)
  }()

  if e.failCmd != nil {
    failError = e.failCmdWrapper(ctx, circuit, execError)

    if failError != nil {
      e.err <- failError
    }
  } else {
    e.err <- execError
  }
}

func (e *executor) execCmdWrapper(ctx context.Context, circuit *circuit) error {
  defer func() {
    if panicErr := recover(); panicErr != nil {
      e.fail(ctx, circuit, fmt.Errorf("exec panic: %s", panicErr))
    }
  }()

  return e.execCmd(ctx)
}

func (e *executor) failCmdWrapper(ctx context.Context, circuit *circuit, err error) error {
  defer func() {
    if panicErr := recover(); panicErr != nil {
      e.err <- fmt.Errorf("failover panic: %s", panicErr)
    }
  }()

  return e.failCmd(ctx, err)
}

func translateError(err error) eventType {
  switch err {
  case errors.ConcurrentLimitError:
    return rejected
  case errors.CancelledError:
    return cancelled
  case errors.TimeoutError:
    return timeout
  }

  return failure
}

func translateFallbackError(err error) eventType {

  if err == nil {
    return fallbackSuccess
  }

  return fallbackFailure
}

package breaker

import (
	"time"
)

const (
	DefaultTimeout = time.Second
	DefaultMaxConcurrentCalls = 1000
	DefaultErrorThreshold = 0.05
	DefaultSleepDuration = time.Second
)

var settings map[string]Settings

func init() {
	settings = make(map[string]Settings)
}

type Settings struct {
	Timeout            time.Duration
	MaxConcurrentCalls int
	ErrorThreshold     float32
	SleepDuration      time.Duration
}

func ConfigureCircuit(name string, s Settings) Settings {
	if s.ErrorThreshold == 0 {
		s.ErrorThreshold = DefaultErrorThreshold
	}

	if s.SleepDuration == 0 {
		s.SleepDuration = DefaultSleepDuration
	}

	if s.MaxConcurrentCalls == 0 {
		s.MaxConcurrentCalls = DefaultMaxConcurrentCalls
	}

	if s.Timeout == 0 {
		s.Timeout =DefaultTimeout
	}

	settings[name] = s
	return s
}

func GetSettings(name string) Settings {
	if s, ok := settings[name]; ok {
		return s
	}

	return Settings {
		DefaultTimeout,
		DefaultMaxConcurrentCalls,
		DefaultErrorThreshold,
		DefaultSleepDuration,
	}
}


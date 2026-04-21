package retry

import (
	"context"
	"errors"
	"time"
)

// Config holds retry configuration.
type Config struct {
	MaxAttempts int           // maximum number of attempts
	Delay       time.Duration // delay between attempts
	MaxDelay    time.Duration // maximum delay (for exponential backoff)
	Multiplier  float64       // backoff multiplier
}

// DefaultConfig returns the default retry configuration.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}
}

// Do executes the function with retry logic.
func Do(fn func() error, cfg ...Config) error {
	config := DefaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	var lastErr error
	delay := config.Delay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 && delay > 0 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if IsNonRetryable(lastErr) {
			return lastErr
		}
	}

	return lastErr
}

// DoWithContext executes the function with retry logic and context cancellation.
func DoWithContext(ctx context.Context, fn func() error, cfg ...Config) error {
	config := DefaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	var lastErr error
	delay := config.Delay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 && delay > 0 {
			select {
			case <-time.After(delay):
				delay = time.Duration(float64(delay) * config.Multiplier)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if IsNonRetryable(lastErr) {
			return lastErr
		}
	}

	return lastErr
}

// IsNonRetryable checks if the error should not be retried.
func IsNonRetryable(err error) bool {
	if err == nil {
		return true
	}
	var nonRetryable *NonRetryableError
	return errors.As(err, &nonRetryable)
}

// NonRetryableError is an error that should not be retried.
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return e.Err.Error()
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

// NewNonRetryableError wraps an error as non-retryable.
func NewNonRetryableError(err error) *NonRetryableError {
	return &NonRetryableError{Err: err}
}

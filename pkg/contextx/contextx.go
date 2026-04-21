package contextx

import (
	"context"
	"time"
)

// WithTimeout returns a new context with timeout.
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// WithDeadline returns a new context with deadline.
func WithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, deadline)
}

// WithCancel returns a new context that can be cancelled.
func WithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}

// WithValue returns a new context with the given key-value pair.
func WithValue(parent context.Context, key, val any) context.Context {
	return context.WithValue(parent, key, val)
}

// Value retrieves a value from the context.
func Value[T any](ctx context.Context, key any) (T, bool) {
	val := ctx.Value(key)
	if val == nil {
		var zero T
		return zero, false
	}
	typed, ok := val.(T)
	return typed, ok
}

// MustValue retrieves a value from the context, panics if not found.
func MustValue[T any](ctx context.Context, key any) T {
	val, ok := Value[T](ctx, key)
	if !ok {
		panic("context value not found: " + any(key).(string))
	}
	return val
}

// Sleep pauses the current goroutine for the specified duration.
func Sleep(d time.Duration) {
	time.Sleep(d)
}

// WaitFor waits for the condition to be true or timeout.
func WaitFor(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// Background returns a non-nil, empty Context.
func Background() context.Context {
	return context.Background()
}

// TODO returns a non-nil, empty Context and a cancel function.
func TODO() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

package resilience

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit open")

const (
	stateClosed = "CLOSED"
	stateOpen   = "OPEN"
	stateHalf   = "HALF_OPEN"
)

type CircuitBreaker struct {
	mu                 sync.Mutex
	state              string
	failures           int
	successes          int
	failureThreshold   int
	halfOpenSuccesses  int
	resetTimeout       time.Duration
	lastFailure        time.Time
}

func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration, halfOpenSuccesses int) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 10 * time.Second
	}
	if halfOpenSuccesses <= 0 {
		halfOpenSuccesses = 1
	}
	return &CircuitBreaker{
		state:             stateClosed,
		failureThreshold:  failureThreshold,
		halfOpenSuccesses: halfOpenSuccesses,
		resetTimeout:      resetTimeout,
	}
}

func (c *CircuitBreaker) Execute(fn func() error) error {
	if err := c.allow(); err != nil {
		return err
	}
	err := fn()
	c.after(err == nil)
	return err
}

func (c *CircuitBreaker) allow() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case stateOpen:
		if time.Since(c.lastFailure) >= c.resetTimeout {
			c.state = stateHalf
			c.failures = 0
			c.successes = 0
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

func (c *CircuitBreaker) after(success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case stateClosed:
		if success {
			c.failures = 0
			return
		}
		c.failures++
		c.lastFailure = time.Now()
		if c.failures >= c.failureThreshold {
			c.state = stateOpen
		}
	case stateHalf:
		if success {
			c.successes++
			if c.successes >= c.halfOpenSuccesses {
				c.state = stateClosed
				c.failures = 0
				c.successes = 0
			}
			return
		}
		c.state = stateOpen
		c.lastFailure = time.Now()
		c.failures = 0
		c.successes = 0
	}
}

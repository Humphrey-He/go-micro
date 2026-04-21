package syncx

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// OnceValue creates a value that is executed only once.
func OnceValue[T any](fn func() T) func() T {
	var once sync.Once
	var val T
	var initialized atomic.Bool
	return func() T {
		if !initialized.Load() {
			once.Do(func() {
				val = fn()
				initialized.Store(true)
			})
		}
		return val
	}
}

// OnceValues creates values that are executed only once.
func OnceValues[T any, V any](fn func() (T, V)) func() (T, V) {
	var once sync.Once
	var val1 T
	var val2 V
	var initialized atomic.Bool
	return func() (T, V) {
		if !initialized.Load() {
			once.Do(func() {
				val1, val2 = fn()
				initialized.Store(true)
			})
		}
		return val1, val2
	}
}

// Pool wraps sync.Pool with generic type.
type Pool[T any] struct {
	pool sync.Pool
}

func NewPool[T any](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFn()
			},
		},
	}
}

// Get retrieves a value from the pool.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns a value to the pool.
func (p *Pool[T]) Put(v T) {
	p.pool.Put(v)
}

// AtomicBool wraps atomic.Bool.
type AtomicBool struct {
	v atomic.Bool
}

// NewAtomicBool creates a new AtomicBool.
func NewAtomicBool(val bool) *AtomicBool {
	b := &AtomicBool{}
	if val {
		b.v.Store(true)
	}
	return b
}

// Load loads the value.
func (b *AtomicBool) Load() bool {
	return b.v.Load()
}

// Store stores the value.
func (b *AtomicBool) Store(val bool) {
	b.v.Store(val)
}

// Swap swaps the value.
func (b *AtomicBool) Swap(val bool) bool {
	return b.v.Swap(val)
}

// CAS compares and swaps the value.
func (b *AtomicBool) CAS(old, new bool) bool {
	return b.v.CompareAndSwap(old, new)
}

// AtomicInt wraps atomic.Int64 for int.
type AtomicInt struct {
	v atomic.Int64
}

// NewAtomicInt creates a new AtomicInt.
func NewAtomicInt(val int) *AtomicInt {
	i := &AtomicInt{}
	i.v.Store(int64(val))
	return i
}

// Load loads the value.
func (i *AtomicInt) Load() int {
	return int(i.v.Load())
}

// Store stores the value.
func (i *AtomicInt) Store(val int) {
	i.v.Store(int64(val))
}

// Add adds delta to the value.
func (i *AtomicInt) Add(delta int) int {
	return int(i.v.Add(int64(delta)))
}

// Inc increments the value.
func (i *AtomicInt) Inc() int {
	return int(i.v.Add(1))
}

// Dec decrements the value.
func (i *AtomicInt) Dec() int {
	return int(i.v.Add(-1))
}

// Swap swaps the value.
func (i *AtomicInt) Swap(val int) int {
	return int(i.v.Swap(int64(val)))
}

// CAS compares and swaps the value.
func (i *AtomicInt) CAS(old, new int) bool {
	return i.v.CompareAndSwap(int64(old), int64(new))
}

// WaitGroup wraps sync.WaitGroup with more methods.
type WaitGroup struct {
	wg sync.WaitGroup
}

// Add adds delta to the WaitGroup.
func (wg *WaitGroup) Add(delta int) {
	wg.wg.Add(delta)
}

// Done signals one task is complete.
func (wg *WaitGroup) Done() {
	wg.wg.Done()
}

// Wait blocks until all tasks are complete.
func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}

// Go runs a function in a goroutine.
func Go(fn func()) {
	go fn()
}

// GoWithRecover runs a function in a goroutine with panic recovery.
func GoWithRecover(recoverFn func(any), fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if recoverFn != nil {
					recoverFn(r)
				}
			}
		}()
		fn()
	}()
}

// Semaphore implements a counting semaphore.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore with the given limit.
func NewSemaphore(limit int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, limit)}
}

// Acquire acquires a permit.
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// Release releases a permit.
func (s *Semaphore) Release() {
	<-s.ch
}

// TryAcquire tries to acquire a permit without blocking.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// TryAcquireWithTimeout tries to acquire a permit with timeout.
func (s *Semaphore) TryAcquireWithTimeout(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	select {
	case s.ch <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// SingleFlight deduplicates concurrent calls.
type SingleFlight struct {
	mu sync.Mutex
	m  map[string]*call[float64]
}

type call[T any] struct {
	wg  sync.WaitGroup
	val T
	err error
}

// Do deduplicates and returns the result of the given function.
func (sf *SingleFlight) Do(key string, fn func() (float64, error)) (float64, error) {
	sf.mu.Lock()
	if sf.m == nil {
		sf.m = make(map[string]*call[float64])
	}
	if c, ok := sf.m[key]; ok {
		sf.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call[float64])
	c.wg.Add(1)
	sf.m[key] = c
	sf.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	sf.mu.Lock()
	delete(sf.m, key)
	sf.mu.Unlock()

	return c.val, c.err
}

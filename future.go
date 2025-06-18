package future

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// Inspired by java.util.concurrent.Future<T> except that we return a tuple instead
// of a proxy object which itself wraps error.
type Future[T any] interface {
	// Get completion value, blocking until another goroutine calls [Set] or [Complete].
	Get() (T, error)
	// Get completion value, with context for cancellation or timeout.
	GetWithContext(context.Context) (T, error)
	// Complete this Future with a result and error.
	// Returns true if this call caused the Future to transition to a completed state.
	Complete(T, error) bool
	// Complete this Future with a cancellation.
	// Returns true if this call caused the Future to transition to a completed state.
	Cancel() bool
	// Returns true if this Future was completed by a call to Cancel().
	IsCancelled() bool
}

type result[T any] struct {
	value T
	err   error
}

type futureImpl[T any] struct {
	result atomic.Pointer[result[T]]
	wg     *sync.WaitGroup
}

func New[T any]() Future[T] {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return &futureImpl[T]{wg: wg}
}

func (f *futureImpl[T]) Get() (T, error) {
	value := f.result.Load()
	if value != nil {
		return value.value, value.err
	}
	f.wg.Wait()
	value = f.result.Load()
	if value == nil {
		panic("empty result after wait")
	}
	return value.value, value.err
}

type token struct{}

func (f *futureImpl[T]) GetWithContext(c context.Context) (T, error) {
	var zero T
	value := f.result.Load()
	if value != nil {
		return value.value, value.err
	}
	waitCh := make(chan token)
	go func() {
		f.wg.Wait()
		close(waitCh)
	}()
	select {
	case <-c.Done():
		return zero, c.Err()
	case <-waitCh:
		value := f.result.Load()
		if value == nil {
			panic("empty result after wait")
		}
		return value.value, value.err
	}
}

func (f *futureImpl[T]) Complete(value T, err error) bool {
	caused := f.result.CompareAndSwap(nil, &result[T]{value: value, err: err})
	if caused {
		f.wg.Done()
	}
	return caused
}

type cancelToken struct{}

var cancelled = cancelToken{}

func (_ cancelToken) Error() string {
	return "future completed via Cancel()"
}
func (f *futureImpl[T]) Cancel() bool {
	caused := f.result.CompareAndSwap(nil, &result[T]{err: cancelled})
	if caused {
		f.wg.Done()
	}
	return caused
}
func (f *futureImpl[T]) IsCancelled() bool {
	value := f.result.Load()
	if value == nil {
		return false
	}
	return errors.Is(value.err, cancelled)
}

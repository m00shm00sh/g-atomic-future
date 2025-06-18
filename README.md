# g-atomic-future

Yet another implementation of futures for Go. This one is different because:
1. We can create futures in an initially incomplete state, at the cost of needing to specify the value type.
2. `GetWithContext(ctx)` for timed (or cancellable) waits.
3. `Complete(value, error)` to explicitly mark completed, like Java's `Future<T>().complete()`.
4. `Cancel()` to complete with cancellation, like Java's `Future<T>().cancel()`.
   - `IsCancelled()` can be used to query if the Future completed as a result of another goroutine calling `Cancel`.
5. Atomic pointer instead of mutex or goroutine writing to channel indefinitely for stored value.

# Example
```go
f := future.New[int]()

func goroutine1() {
    r, e := f.Get()
}

func goroutine2() {
    f.Complete(1, nil)
}

func goroutine3() {
    f.Cancel()
}
```
`goroutine1` will wait for either of `goroutine2` or `goroutine3` to finish but subsequent `Get()`s will be non-blocking.


Timed wait:
```go
f := future.New[int]()

func goroutine1() {
	ctx, cancel := context.WithTimeout(context.Background(), 1000 * time.Millisecond)
	defer cancel()
    r, e := f.GetWithContext(ctx)
}

func goroutine2() {
    f.Complete(1, nil)
}

func goroutine3() {
    f.Cancel()
}
```
`goroutine1` will wait for either of `goroutine2` or `goroutine3` to finish or timeout to occur
but subsequent `Get()`s will be non-blocking.


package future

import (
	"context"
	"errors"
	"testing"
	"time"
)

var eTimeout = errors.New("timeout")

func TestSetGet(t *testing.T) {
	f := New[int]()
	e := errors.New("")
	if r := timedTest[bool](t, 100 * time.Millisecond, func () bool { return f.Complete(1, e) }); r == nil || *r != true {
		t.Error("completion not triggered")
		return
	}
	if r, _ := timedTestE[int](t, 100 * time.Millisecond, func () (int, error) { return f.Get() }); r == nil || *r != 1 {
		t.Error("completion not triggered")
		return
	}
}

func TestSetGetcontext(t *testing.T) {
	f := New[int]()
	e := errors.New("")
	if r := timedTest[bool](t, 100 * time.Millisecond, func () bool { return f.Complete(1, e) }); r == nil || *r != true {
		t.Error("completion not triggered")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1000 * time.Millisecond)
	defer cancel()
	if r, _ := timedTestE[int](t, 100 * time.Millisecond, func () (int, error) { return f.GetWithContext(ctx) }); r == nil || *r != 1 {
		t.Error("completion not triggered")
		return
	}
}

func testGetcontext(t *testing.T) {
	f := New[int]()
	ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Millisecond)
	defer cancel()
	if _, e := timedTestE[int](t, 1000 * time.Millisecond, func () (int, error) { return f.GetWithContext(ctx) }); e != eTimeout {
		t.Error("context not triggered")
		return
	}
}

func TestCancelGet(t *testing.T) {
	f := New[int]()
	if r := timedTest[bool](t, 100 * time.Millisecond, func () bool { return f.Cancel() }); r == nil || *r != true {
		t.Error("completion not triggered")
		return
	}
	if !f.IsCancelled() {
		t.Error("completion not triggered")
		return
	}
}

func timedTest[R any](t *testing.T, timeout time.Duration, callable func () R) *R {
	ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Millisecond)
	waitCh := make(chan R)
	defer cancel()
	go func() {
		waitCh <- callable()
		close(waitCh)
	}()
	select {
	case <- ctx.Done():
		t.Error("timeout")
		return nil
	case r := <- waitCh:
		return &r
	}
	panic("unreachable")
}
	
func timedTestE[R any](t *testing.T, timeout time.Duration, callable func () (R, error)) (*R, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Millisecond)
	type wRes struct {
		r R
		e error
	}
	waitCh := make(chan wRes)
	defer cancel()
	go func() {
		r, e := callable()
		waitCh <- wRes{r,e}
		close(waitCh)
	}()
	select {
	case <- ctx.Done():
		t.Error("timeout")
		return nil, eTimeout
	case re := <- waitCh:
		r := re.r
		e := re.e
		return &r, e
	}
	panic("unreachable")
}
	

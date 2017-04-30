package lifecycle_test

import (
	"context"
	"testing"
	"time"

	example "github.com/boz/go-lifecycle/_example"
)

func TestLifecycle_shutdown(t *testing.T) {
	cache := example.NewCache(context.Background())
	runTestWithShutdown(t, cache, cache.Shutdown, "cache.Shutdown()")
}

func TestLifecycle_ctx_cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache := example.NewCache(ctx)
	runTestWithShutdown(t, cache, cancel, "context.Cancel()")
}

func runTestWithShutdown(t *testing.T, cache example.Cache, stopfn func(), msg string) {
	if err := cache.Put("foo", "bar"); err != nil {
		t.Errorf("%v: unable to put before shutdown: %v", msg, err)
	}

	if _, err := cache.Get("foo"); err != nil {
		t.Errorf("%v: unable to get before shutdown: %v", msg, err)
	}

	select {
	case <-cache.Done():
		t.Error("%v: done readable before shutdown", msg)
	default:
	}

	stopfn()

	select {
	case <-cache.Done():
	case <-time.After(time.Millisecond * 10):
		t.Errorf("%v: shutdown not completed after 10ms")
	}

	if err := cache.Put("foo", "bar"); err != example.ErrNotRunning {
		t.Errorf("%v: invalid err after shutdown: %v", msg, err)
	}

	if _, err := cache.Get("foo"); err != example.ErrNotRunning {
		t.Errorf("%v: invalid err after shutdown: %v", msg, err)
	}

}

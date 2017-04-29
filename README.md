# go-lifecycle [![Build Status](https://travis-ci.org/boz/go-lifecycle.svg?branch=master)](https://travis-ci.org/boz/go-lifecycle)

Lifecycle utility for goroutine-based objects.

See [here](_example/example.go) for a complete example.

## Example

```go
// ... 
type cache struct {
	lc    lifecycle.Lifecycle
  // ...
}

// ...

func NewCache(ctx context.Context) Cache {
	c := &cache{
		lc:    lifecycle.New(),
    // ...
	}
	go c.lc.WatchContext(ctx)
	go c.run()
	return c
}

func (c *cache) run() {
	defer c.lc.ShutdownCompleted()
	var drainedch chan bool
	stopch := c.lc.ShutdownRequest()

	for {
		select {
		case <-stopch:
			stopch = nil
			c.lc.ShutdownInitiated()

			// simulate any necessary draining
			drainedch = make(chan bool, 1)
			drainedch <- true

		case <-drainedch:
			return
    // ...
		}
	}
}

func (c *cache) Put(key, value string) error {
	select {
	case c.putch <- putreq{key, value}:
		return nil
	case <-c.lc.ShuttingDown():
		return ErrNotRunning
	}
}

func (c *cache) Get(key string) (string, error) {
	valch := make(chan string)

	select {
	case c.getch <- getreq{key, valch}:
		return <-valch, nil
	case <-c.lc.ShuttingDown():
		return "", ErrNotRunning
	}
}

func (c *cache) Shutdown() {
	c.lc.Shutdown()
}

func (c *cache) Done() <-chan struct{} {
	return c.lc.Done()
}

```

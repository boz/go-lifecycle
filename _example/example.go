package example

import (
	"context"
	"fmt"

	lifecycle "github.com/boz/go-lifecycle"
)

var (
	ErrNotRunning = fmt.Errorf("not running")
)

type Cache interface {
	Put(string, string) error
	Get(string) (string, error)
	Shutdown()
	Done() <-chan struct{}
}

type cache struct {
	lc    lifecycle.Lifecycle
	putch chan putreq
	getch chan getreq
	items map[string]string
}

type putreq struct {
	key   string
	value string
}

type getreq struct {
	key   string
	valch chan<- string
}

func NewCache(ctx context.Context) Cache {
	c := &cache{
		lc:    lifecycle.New(),
		putch: make(chan putreq),
		getch: make(chan getreq),
		items: make(map[string]string),
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
		case req := <-c.putch:
			c.items[req.key] = req.value
		case req := <-c.getch:
			req.valch <- c.items[req.key]
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

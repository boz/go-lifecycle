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
	Shutdown(error)
	ShutdownAsync(error)
	Done() <-chan struct{}
	Error() error
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

loop:
	for {
		select {
		case err := <-c.lc.ShutdownRequest():
			c.lc.ShutdownInitiated(err)
			break loop
		case req := <-c.putch:
			c.items[req.key] = req.value
		case req := <-c.getch:
			req.valch <- c.items[req.key]
		}
	}

	// shutdown, wait for child resources
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

func (c *cache) Shutdown(err error) {
	c.lc.Shutdown(err)
}

func (c *cache) ShutdownAsync(err error) {
	c.lc.ShutdownAsync(err)
}

func (c *cache) Done() <-chan struct{} {
	return c.lc.Done()
}

func (c *cache) Error() error {
	return c.lc.Error()
}

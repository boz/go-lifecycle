package lifecycle

import "context"

type Lifecycle interface {
	LifecycleReader
	ShutdownRequest() <-chan struct{}
	ShutdownInitiated()
	ShutdownCompleted()
	WatchContext(context.Context)
	Shutdown()
}

type LifecycleReader interface {
	ShuttingDown() <-chan struct{}
	Done() <-chan struct{}
}

type lifecycle struct {
	stopch     chan struct{}
	stoppingch chan struct{}
	stoppedch  chan struct{}
}

func New() Lifecycle {
	return &lifecycle{
		stopch:     make(chan struct{}),
		stoppingch: make(chan struct{}),
		stoppedch:  make(chan struct{}),
	}
}

// ShutdownRequest() returns a channel that is available for reading when
// a shutdown has requested.
func (l *lifecycle) ShutdownRequest() <-chan struct{} {
	return l.stopch
}

// ShutdownInitiated() declares that shutdown has begun.  Will panic if called twice.
func (l *lifecycle) ShutdownInitiated() {
	close(l.stoppingch)
}

// ShuttingDown() returns a channel that is available for reading
// after ShutdownInitiated() has been called.
func (l *lifecycle) ShuttingDown() <-chan struct{} {
	return l.stoppingch
}

// ShutdownCompleted() declares that shutdown has completed.  Will panic if called twice.
func (l *lifecycle) ShutdownCompleted() {
	close(l.stoppedch)
}

// Done() returns a channel that is available for reading
// after ShutdownCompleted() has been called.
func (l *lifecycle) Done() <-chan struct{} {
	return l.stoppedch
}

// Shutdown() initiates shutdown by sending a value to the channel
// requtned by ShutdownRequest() and blocks untill ShutdownCompleted()
// is called.
func (l *lifecycle) Shutdown() {
	for {
		select {
		case l.stopch <- struct{}{}:
		case <-l.stoppingch:
			<-l.stoppedch
			return
		}
	}
}

// WatchContext() observes the given context and initiates a shutdown
// if the context is shutdown before the lifecycle is.
func (l *lifecycle) WatchContext(ctx context.Context) {
	var stopch chan struct{}
	donech := ctx.Done()
	for {
		select {
		case <-l.stoppingch:
			return
		case <-donech:
			donech = nil
			stopch = l.stopch
		case stopch <- struct{}{}:
			return
		}
	}
}

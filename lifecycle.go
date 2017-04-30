package lifecycle

import "context"

type Lifecycle interface {
	LifecycleReader

	// ShutdownRequest() returns a channel that is available for reading when
	// a shutdown has requested.
	ShutdownRequest() <-chan struct{}

	// ShutdownInitiated() declares that shutdown has begun.  Will panic if called twice.
	ShutdownInitiated()

	// ShutdownCompleted() declares that shutdown has completed.  Will panic if called twice.
	ShutdownCompleted()

	// WatchContext() observes the given context and initiates a shutdown
	// if the context is shutdown before the lifecycle is.
	WatchContext(context.Context)

	// Shutdown() initiates shutdown by sending a value to the channel
	// requtned by ShutdownRequest() and blocks untill ShutdownCompleted()
	// is called.
	Shutdown()
}

// LifecycleReader exposes read-only access to lifecycle state.
type LifecycleReader interface {
	// ShuttingDown() returns a channel that is available for reading
	// after ShutdownInitiated() has been called.
	ShuttingDown() <-chan struct{}

	// Done() returns a channel that is available for reading
	// after ShutdownCompleted() has been called.
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

func (l *lifecycle) ShutdownRequest() <-chan struct{} {
	return l.stopch
}

func (l *lifecycle) ShutdownInitiated() {
	close(l.stoppingch)
}

func (l *lifecycle) ShuttingDown() <-chan struct{} {
	return l.stoppingch
}

func (l *lifecycle) ShutdownCompleted() {
	close(l.stoppedch)
}

func (l *lifecycle) Done() <-chan struct{} {
	return l.stoppedch
}

func (l *lifecycle) Shutdown() {
	select {
	case <-l.stoppedch:
		return
	case l.stopch <- struct{}{}:
	case <-l.stoppingch:
	}
	<-l.stoppedch
}

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

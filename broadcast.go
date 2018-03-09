package luigi // import "cryptoscope.co/go/luigi"

import (
	"context"
	"sync"
)

type Broadcast interface {
	Register(Sink) func()
}

func NewBroadcast() (Sink, Broadcast) {
	bcst := broadcast{sinks: make(map[*Sink]struct{})}

	return (*broadcastSink)(&bcst), &bcst
}

type broadcast struct {
	sync.Mutex
	sinks map[*Sink]struct{}
}

func (bcst *broadcast) Register(sink Sink) func() {
	bcst.Lock()
	defer bcst.Unlock()
	bcst.sinks[&sink] = struct{}{}

	return func() {
		bcst.Lock()
		defer bcst.Unlock()
		delete(bcst.sinks, &sink)
	}
}

type broadcastSink broadcast

func (bcst *broadcastSink) Pour(ctx context.Context, v interface{}) error {
	var wg sync.WaitGroup

	bcst.Lock()
	defer bcst.Unlock()

	n := len(bcst.sinks)
	wg.Add(n)

	errCh := make(chan error)
	errOut := make(chan error, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		defer close(errOut)

		if err, ok := <-errCh; ok {
			cancel()
			errOut <- err
		}

		for range errCh {
			// drain
		}
	}()

	for sinkptr := range bcst.sinks {
		go func(sink Sink) {
			defer wg.Done()

			err := sink.Pour(ctx, v)
			if err != nil {
				errCh <- err
				return
			}
		}(*sinkptr)
	}

	wg.Wait()
	close(errCh)

	return <-errOut
}

func (bcst *broadcastSink) Close() error { return nil }

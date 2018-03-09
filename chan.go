package luigi // import "cryptoscope.co/go/luigi"

import "context"

type chanSource struct {
	ch <-chan interface{}
}

func (src *chanSource) Next(ctx context.Context) (v interface{}, err error) {
	var ok bool

	select {
	case v, ok = <-src.ch:
		if !ok {
			err = EOS{}
		}
	case <-ctx.Done():
		err = ctx.Err()
	}

	return v, err
}

type chanSink struct {
	ch chan<- interface{}
}

func (sink *chanSink) Pour(ctx context.Context, v interface{}) error {
	select {
	case sink.ch <- v:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (sink *chanSink) Close() error {
	close(sink.ch)
	return nil
}

func NewPipe() (Source, Sink) {
	ch := make(chan interface{})

	return &chanSource{ch}, &chanSink{ch}
}

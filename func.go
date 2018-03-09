package luigi // import "cryptoscope.co/go/luigi"

import (
	"context"
)

type funcSink func(context.Context, interface{}, bool) error

func (fSink funcSink) Pour(ctx context.Context, v interface{}) error {
	return fSink(ctx, v, false)
}

func (fSink funcSink) Close() error {
	return fSink(nil, nil, true)
}

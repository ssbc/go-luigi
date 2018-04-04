package luigi // import "cryptoscope.co/go/luigi"

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestBufferedPipe(t *testing.T) {
	r := require.New(t)
	src, sink := NewPipe(WithBuffer(2))

	ctx := context.Background()
	err := sink.Pour(ctx, 1)
	r.NoError(err, "pouring first value")
	err = sink.Pour(ctx, 2)
	r.NoError(err, "pouring second value")
	err = sink.Close()
	r.NoError(err, "closing sink")

	v, err := src.Next(ctx)
	r.NoError(err, "getting first value")
	r.Equal(v, 1, "expected 1")

	v, err = src.Next(ctx)
	r.NoError(err, "getting first value")
	r.Equal(v, 2, "expected 1")

	_, err = src.Next(ctx)
	r.Equal(errors.Cause(err), EOS{}, "expected end-of-stream")
}

func TestCloseWithError(t *testing.T) {
	r := require.New(t)
	src, sink := NewPipe(WithBuffer(2))

	ctx := context.Background()

	err := sink.Pour(ctx, 1)
	r.NoError(err, "pouring first value")
	err = sink.Pour(ctx, 2)
	r.NoError(err, "pouring second value")
	err = sink.(ErrorCloser).CloseWithError(errors.New("an unfortunate error occurred"))
	r.NoError(err, "closing sink")

	v, err := src.Next(ctx)
	r.NoError(err, "getting first value")
	r.Equal(v, 1, "expected 1")

	v, err = src.Next(ctx)
	r.NoError(err, "getting first value")
	r.Equal(v, 2, "expected 1")

	_, err = src.Next(ctx)
	r.Equal("an unfortunate error occurred", errors.Cause(err).Error(), "expected custom error")
}

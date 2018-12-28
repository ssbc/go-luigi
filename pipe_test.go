package luigi // import "go.cryptoscope.co/luigi"

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestBufferedPipe(t *testing.T) {
	r := require.New(t)

	mkTest := func(bufSize int) func(*testing.T) {
		return func(t *testing.T) {
			src, sink := NewPipe(WithBuffer(bufSize))

			ctx := context.Background()

			for i := 0; i < bufSize; i++ {
				err := sink.Pour(ctx, i)
				r.NoErrorf(err, "pouring %d", i)
			}

			err := sink.Close()
			r.NoError(err, "closing")

			for i := 0; i < bufSize; i++ {
				v, err := src.Next(ctx)
				r.NoErrorf(err, "getting what I expect to be %d", i)
				r.Equalf(v, i, "unexpected value %v, got %v", v, i)
			}

			_, err = src.Next(ctx)
			r.Equal(EOS{}, errors.Cause(err), "expected end-of-stream")
		}
	}

	sizes := func() []int {
		var (
			acc int
			out = make([]int, 16)
		)

		for i := range out {
			out[i] = acc
			if acc == 0 {
				acc++
			} else {
				acc *= 2
			}
		}

		return out
	}()

	for _, size := range sizes {
		t.Run(fmt.Sprint(size), mkTest(size))
	}
}

func TestCloseWithError(t *testing.T) {
	errMsg := "an unfortunate error occurred"

	r := require.New(t)
	src, sink := NewPipe(WithBuffer(2))

	ctx := context.Background()

	err := sink.Pour(ctx, 1)
	r.NoError(err, "pouring first value")
	err = sink.Pour(ctx, 2)
	r.NoError(err, "pouring second value")
	err = sink.(ErrorCloser).CloseWithError(errors.New(errMsg))
	r.NoError(err, "closing sink")

	v, err := src.Next(ctx)
	r.NoError(err, "getting first value")
	r.Equal(v, 1, "expected 1")

	v, err = src.Next(ctx)
	r.NoError(err, "getting first value")
	r.Equal(v, 2, "expected 1")

	_, err = src.Next(ctx)
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}

	r.Equal(errMsg, errors.Cause(err).Error(), "expected custom error")
}

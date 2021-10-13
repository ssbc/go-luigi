// SPDX-FileCopyrightText: 2021 The Luigi Authors
//
// SPDX-License-Identifier: MIT

package luigi

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// verify we can abort a pipe while it is waiting to read/write it's data

func TestCloseWhileNext(t *testing.T) {
	ctx := context.TODO()
	r := require.New(t)

	errc := make(chan error)

	src, sink := NewPipe()

	go func() {
		errc <- sink.Close()
	}()

	v, err := src.Next(ctx)
	r.Equal(EOS{}, err, "should return end of stream")
	r.Nil(v)

	r.NoError(<-errc)
}

func TestCloseWhilePour(t *testing.T) {
	ctx := context.TODO()
	r := require.New(t)

	closeErr := make(chan error)
	nextErr := make(chan error)

	src, sink := NewPipe()

	go func() {
		time.Sleep(25 * time.Millisecond) // just make sure Pour() started
		closeErr <- sink.Close()
		go func() {
			v, err := src.Next(ctx)
			if v != nil {
				nextErr <- errors.Errorf("expected nil value but got %v", v)
			}
			if !IsEOS(err) {
				nextErr <- errors.Wrap(err, "error from next")
			}
			close(nextErr)
		}()
	}()

	err := sink.Pour(ctx, "test msg")
	// TODO use exported error variables
	r.EqualError(err, "pour to closed sink", "should return end of stream")

	r.NoError(<-closeErr)

	for err := range nextErr {
		r.NoError(err)
	}
}

func TestNextWhileCanceld(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	r := require.New(t)

	errc := make(chan error)

	src, sink := NewPipe()

	wait := make(chan struct{})
	go func() {
		close(wait)
	}()

	go func() {
		<-wait
		cancel()
		errc <- sink.Close()
	}()

	<-wait
	v, err := src.Next(ctx)
	r.NotNil(err, "has to return an error")
	r.Nil(v)

	r.NoError(<-errc)
}

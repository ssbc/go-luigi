package luigi // import "cryptoscope.co/go/luigi"

import (
	"context"
	"testing"
	"time"
)

func TestChanSource(t *testing.T) {
	type testcase struct {
		values  []interface{}
		doClose bool
	}

	test := func(tc testcase) {
		ch := make(chan interface{})
		cs := &chanSource{ch: ch, nonBlocking: false}

		for _, v := range tc.values {
			go func(v_ interface{}) {
				ch <- v_
			}(v)

			v_, err := cs.Next(context.TODO())

			if v != v_ {
				t.Errorf("expected value %#v, but got %#v", v, v_)
			}

			if err != nil {
				t.Errorf("expected nil error but got %s", err)
			}
		}

		if tc.doClose {
			close(ch)
			_, err := cs.Next(context.TODO())
			if !IsEOS(err) {
				t.Errorf("expected end-of-stream error but got %s", err)
			}
		} else {
			ctx, cancel := context.WithTimeout(
				context.Background(), 5*time.Millisecond)
			defer cancel()

			_, err := cs.Next(ctx)
			if err != context.DeadlineExceeded {
				t.Errorf("expected deadline exceeded error, got %v", err)
			}
		}
	}

	cases := []testcase{
		{[]interface{}{1, 2, 3}, true},
		{[]interface{}{}, true},
		{[]interface{}{nil, 0, ""}, false},
	}

	for _, tc := range cases {
		test(tc)
	}
}

func TestChanSink(t *testing.T) {
	type testcase struct {
		values []interface{}
	}

	test := func(tc testcase) {
		ch := make(chan interface{})
		echoCh := make(chan interface{})
		cs := &chanSink{ch: ch, nonBlocking: false}

		for _, v := range tc.values {
			go func() {
				echoCh <- (<-ch)
			}()

			err := cs.Pour(context.TODO(), v)

			if err != nil {
				t.Errorf("expected nil error but got %s", err)
				break
			}

			v_ := <-echoCh
			if v != v_ {
				t.Errorf("expected value %#v, but got %#v", v, v_)
			}

		}
		go func() {
			_, closed := <-ch
			if !closed {
				echoCh <- nil
			} else {
				close(echoCh)
			}
		}()

		cs.Close()
		_, closed := <-echoCh
		if !closed {
			t.Error("expected closed channel, but read was successful")
		}
	}

	cases := []testcase{
		{[]interface{}{1, 2, 3}},
		{[]interface{}{}},
		{[]interface{}{nil, 0, ""}},
	}

	for _, tc := range cases {
		test(tc)
	}
}

func TestPipe(t *testing.T) {
	type testcase struct {
		values  []interface{}
		doClose bool
	}

	test := func(tc testcase) {
		src, sink := NewPipe()

		errCh := make(chan error)

		for _, v := range tc.values {
			go func(v_ interface{}) {
				errCh <- sink.Pour(context.TODO(), v_)
			}(v)

			v_, err := src.Next(context.TODO())

			if v != v_ {
				t.Errorf("expected value %#v, but got %#v", v, v_)
			}

			if err != nil {
				t.Errorf("expected nil error but got %s", err)
			}

			err = <-errCh
			if err != nil {
				t.Errorf("expected nil error but got %s", err)
			}
		}

		if tc.doClose {
			err := sink.Close()
			if err != nil {
				t.Errorf("sink close: expected nil error, got %v", err)
			}

			_, err = src.Next(context.TODO())
			if !IsEOS(err) {
				t.Errorf("expected end-of-stream error but got %s", err)
			}
		} else {
			ctx, cancel := context.WithTimeout(
				context.Background(), 5*time.Millisecond)
			defer cancel()

			_, err := src.Next(ctx)
			if err != context.DeadlineExceeded {
				t.Errorf("expected deadline exceeded error, got %v", err)
			}
		}
	}

	cases := []testcase{
		{[]interface{}{1, 2, 3}, true},
		{[]interface{}{}, true},
		{[]interface{}{nil, 0, ""}, false},
	}

	for _, tc := range cases {
		test(tc)
	}

}

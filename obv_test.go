package luigi // import "go.cryptoscope.co/luigi"

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestObservable(t *testing.T) {
	type testcase struct {
		values []interface{}
	}

	test := func(tc testcase) {
		obv := NewObservable(nil)

		makeSink := func() (Sink, <-chan interface{}) {
			vChan := make(chan interface{}, 1)
			var (
				lock   sync.Mutex
				closed bool
				first  bool = true
			)
			var sink Sink = FuncSink(func(ctx context.Context, v interface{}, doClose bool) error {
				lock.Lock()
				defer lock.Unlock()

				if first {
					first = false
					return nil
				}

				if closed {
					return fmt.Errorf("call on closed sink")
				}

				if doClose {
					closed = true
					close(vChan)
					return nil
				}

				vChan <- v
				return nil
			})

			return sink, vChan
		}

		perstSink, perstChan := makeSink()
		perstCancel := obv.Register(perstSink)
		defer perstCancel()

		for _, v := range tc.values {
			// use closure so we can defer inside for loop
			func() {
				ephSink, ephChan := makeSink()
				ephCancel := obv.Register(ephSink)
				defer ephCancel()

				obv.Set(v)

				if v_, err := obv.Value(); err != nil {
					t.Errorf("expected error nil, got %v", err)
				} else if v_ != v {
					t.Errorf("expected %v, got %v", v, v_)
				}

				if v_ := <-ephChan; v_ != v {
					t.Errorf("expected %v, got %v", v, v_)
				}

				if v_ := <-perstChan; v_ != v {
					t.Errorf("expected %v, got %v", v, v_)
				}
			}()
		}
	}

	tcs := []testcase{
		{[]interface{}{1, 2, 3}},
	}

	for _, tc := range tcs {
		test(tc)
	}
}

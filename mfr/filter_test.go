package mfr // import "cryptoscope.co/go/luigi/mfr"

import (
	"context"
	"fmt"
	"testing"

	"cryptoscope.co/go/luigi"
)

func TestFilterSink(t *testing.T) {
	type testcase struct {
		in, out    []interface{}
		errStrings []string
	}

	mkTest := func(tc testcase) func(*testing.T) {
		return func(t *testing.T) {
			var (
				iCheck int
				closed bool
				check  luigi.FuncSink = func(ctx context.Context, v interface{}, doClose bool) error {
					defer func() { iCheck++ }()
					if closed && iCheck < len(tc.out) {
						//return fmt.Errorf("received close, but there are values left (i:%v, v:%v, len(out):%v",
						//  iCheck, v, len(tc.out))

						// incoming values after close, ignore so we can test this
						return nil
					} else if doClose {
						closed = true
						return nil
					} else if iCheck >= len(tc.out) {
						return fmt.Errorf("received more values than expected (i:%v, v:%v, len(out):%v",
							iCheck, v, len(tc.out))
					}

					if v != tc.out[iCheck] {
						return fmt.Errorf("expected value %v, but got %v (i=%v)", tc.out[iCheck], v, iCheck)
					}

					return nil
				}
			)

			sf := SinkFilter(check, func(ctx context.Context, v interface{}) (bool, error) {
				vInt, ok := v.(int)
				if !ok {
					return false, TypeError{expected: vInt, got: v}
				}
				return vInt%2 == 0, nil
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for i, v := range tc.in {
				err := sf.Pour(ctx, v)
				if tc.errStrings[i] == "" && err != nil {
					t.Errorf("unexpected pour error: %v", err)
				} else if tc.errStrings[i] != "" && err == nil {
					t.Errorf("expected error %q but got nil", tc.errStrings[i])
				} else if err != nil && err.Error() != tc.errStrings[i] {
					t.Errorf("expected error %q but got: %v", tc.errStrings[i], err)
				}
			}

			if err := sf.Close(); err != nil {
				t.Errorf("error closing sink: %v", err)
			}
		}
	}

	tcs := []testcase{
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{2, 10, 18},
			errStrings: []string{"", "", "", "", "", "", "", ""},
		},
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{2, 10, 18, nil},
			errStrings: []string{"", "", "", "", "", "", "", "", "end of stream"},
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), mkTest(tc))
	}
}

func TestFilterSource(t *testing.T) {
	type testcase struct {
		in, out    []interface{}
		errStrings []string
	}

	mkTest := func(tc testcase) func(*testing.T) {
		return func(t *testing.T) {
			var (
				iIn int
				src luigi.FuncSource = func(ctx context.Context) (interface{}, error) {
					if iIn >= len(tc.in) {
						return nil, luigi.EOS{}
					}

					defer func() { iIn++ }()

					return tc.in[iIn], nil
				}
			)

			srcf := SourceFilter(src, func(ctx context.Context, v interface{}) (bool, error) {
				vInt, ok := v.(int)
				if !ok {
					return false, TypeError{expected: vInt, got: v}
				}
				return vInt%2 == 0, nil
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var (
				err error
				v   interface{}
			)
			for i, exp := range tc.out {
				v, err = srcf.Next(ctx)
				if err != nil && tc.errStrings[i] == err.Error() {
					t.Errorf("unexpected error in call to Next: %v - expected %q", err, tc.errStrings[i])
				} else if err == nil && tc.errStrings[i] != "" {
					t.Errorf("unexpected nil error in call to Next, expected: %s", tc.errStrings[i])
				}

				if v != exp {
					t.Errorf("expected %v, got %v", exp, v)
				}

				i++
			}

			_, err = srcf.Next(ctx)
			if !luigi.IsEOS(err) {
				t.Errorf("expected end-of-stream, got %v", err)
			}
		}
	}

	tcs := []testcase{
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{2, 10, 18},
			errStrings: []string{"", "", "", "", "", "", "", ""},
		},
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{2, 10, 18, nil},
			errStrings: []string{"", "", "", "", "", "", "", "", "end of stream"},
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), mkTest(tc))
	}
}

/* TODO delete this block once it is in version control
func TestFilterSourcePiped(t *testing.T) {
	type testcase struct {
		in, out    []interface{}
		errStrings []string
	}

	mkTest := func(tc testcase) func(*testing.T) {
		return func(t *testing.T) {
			src, sink := luigi.NewPipe()

			srcf := SourceFilter(src, func(ctx context.Context, v interface{}) (bool, error) {
				vInt, ok := v.(int)
				if !ok {
					return false, TypeError{expected: vInt, got: v}
				}
				return vInt%2 == 0, nil
			})

			wait := make(chan struct{})

			go func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var (
					i   int
					v   interface{}
					err error
				)
				for {
					v, err = srcf.Next(ctx)
					if err != nil {
						break
					}

					if i >= len(tc.out) {
						t.Errorf("expected abort, read %v at i=%v", v, i)
					}

					if v != tc.out[i] {
						t.Errorf("expected %v, got %v", tc.out[i], v)
					}

					i++
				}

				if !luigi.IsEOS(err) {
					t.Errorf("expected end-of-stream, got %v", err)
				}
				close(wait)
			}()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for i, v := range tc.in {
				err := sink.Pour(ctx, v)
				if tc.errStrings[i] == "" && err != nil {
					t.Errorf("unexpected pour error: %v", err)
				} else if tc.errStrings[i] != "" && err == nil {
					t.Errorf("expected error %q but got nil", tc.errStrings[i])
				} else if err != nil && err.Error() != tc.errStrings[i] {
					t.Errorf("expected error %q but got: %v", tc.errStrings[i], err)
				}
			}

			err := sink.Close()
			if err != nil {
				t.Errorf("error closing sink: %v", err)
			}

			<-wait
		}
	}

	tcs := []testcase{
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{2, 10, 18},
			errStrings: []string{"", "", "", "", "", "", "", ""},
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), mkTest(tc))
	}
}
*/

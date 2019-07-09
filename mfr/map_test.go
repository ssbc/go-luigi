package mfr // import "go.cryptoscope.co/luigi/mfr"

import (
	"context"
	"fmt"
	"testing"

	"go.cryptoscope.co/luigi"
)

func ExampleSourceMap() {
	toRune := func(
		_ context.Context,
		v interface{},
	) (interface{}, error) {
		return rune(v.(int) + 97), nil
	}

	numbers := luigi.SliceSource([]interface{}{0, 1, 2, 3, 4})
	runes := SourceMap(&numbers, toRune)

	for {
		v, err := runes.Next(context.Background())
		if luigi.IsEOS(err) {
			break
		}
		fmt.Print(string(v.(rune)))
	}
	// Output: abcde
}

func TestMapSink(t *testing.T) {
	type testcase struct {
		in, out    []interface{}
		errStrings []string
	}

	mkTest := func(tc testcase) func(*testing.T) {
		return func(t *testing.T) {
			var (
				iCheck int
				closed bool
				check  luigi.FuncSink = func(ctx context.Context, v interface{}, err error) error {
					defer func() { iCheck++ }()
					if closed && iCheck < len(tc.out) {
						//return fmt.Errorf("received close, but there are values left (i:%v, v:%v, len(out):%v",
						//  iCheck, v, len(tc.out))

						// incoming values after close, ignore so we can test this
						return nil
					} else if err != nil {
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

			sm := SinkMap(check, func(ctx context.Context, v interface{}) (interface{}, error) {
				vInt, ok := v.(int)
				if !ok {
					return false, TypeError{expected: vInt, got: v}
				}
				return vInt % 7, nil
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for i, v := range tc.in {
				err := sm.Pour(ctx, v)
				if tc.errStrings[i] == "" && err != nil {
					t.Errorf("unexpected pour error: %v", err)
				} else if tc.errStrings[i] != "" && err == nil {
					t.Errorf("expected error %q but got nil", tc.errStrings[i])
				} else if err != nil && err.Error() != tc.errStrings[i] {
					t.Errorf("expected error %q but got: %v", tc.errStrings[i], err)
				}
			}

			if err := sm.Close(); err != nil {
				t.Errorf("error closing sink: %v", err)
			}
		}
	}

	tcs := []testcase{
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{1, 2, 5, 0, 3, 4, 0, 2},
			errStrings: []string{"", "", "", "", "", "", "", ""},
		},
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{1, 2, 5, 0, 3, 4, 0, 2, nil},
			errStrings: []string{"", "", "", "", "", "", "", "", "end of stream"},
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), mkTest(tc))
	}
}

func TestMapSource(t *testing.T) {
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

			srcm := SourceMap(src, func(ctx context.Context, v interface{}) (interface{}, error) {
				vInt, ok := v.(int)
				if !ok {
					return false, TypeError{expected: vInt, got: v}
				}
				return vInt % 7, nil
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var (
				err error
				v   interface{}
			)
			for i, exp := range tc.out {
				v, err = srcm.Next(ctx)
				if err != nil && tc.errStrings[i] != err.Error() {
					t.Errorf("unexpected error in call to Next: %q - expected %q", err.Error(), tc.errStrings[i])
				} else if err == nil && tc.errStrings[i] != "" {
					t.Errorf("unexpected nil error in call to Next, expected: %s", tc.errStrings[i])
				}

				if v != exp {
					t.Errorf("expected %v, got %v", exp, v)
				}

				i++
			}

			_, err = srcm.Next(ctx)
			if !luigi.IsEOS(err) {
				t.Errorf("expected end-of-stream, got %v", err)
			}
		}
	}

	tcs := []testcase{
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{1, 2, 5, 0, 3, 4, 0, 2},
			errStrings: []string{"", "", "", "", "", "", "", ""},
		},
		{
			in:         []interface{}{1, 2, 5, 7, 10, 18, 21, 23},
			out:        []interface{}{1, 2, 5, 0, 3, 4, 0, 2, nil},
			errStrings: []string{"", "", "", "", "", "", "", "", "end of stream"},
		},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), mkTest(tc))
	}
}

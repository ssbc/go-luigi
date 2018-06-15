package json // import "go.cryptoscope.co/luigi/json"

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"reflect"
	"testing"

	"go.cryptoscope.co/luigi"
)

func TestSource(t *testing.T) {
	type jsonType struct {
		K       string        `json:"k"`
		Answer  int           `json:"answer"`
		Correct bool          `json:"correct"`
		Tags    []interface{} `json:"tags"`
	}

	type testcase struct {
		jsonString string
		tipe       interface{}
		results    []interface{}
		errchk     func(error) error
	}

	test := func(tc testcase) {
		buf := bytes.NewBuffer([]byte(tc.jsonString))
		src := NewSource(buf, tc.tipe)

		var (
			v   interface{}
			err error
			i   int
		)

		for {
			v, err = src.Next(context.TODO())
			if err = tc.errchk(err); err != nil {
				if luigi.IsEOS(err) {
					if i != len(tc.results) {
						t.Error(err)
					}
					break
				}

				t.Errorf("error:%v; tc:%v; i:%v", err, tc, i)
				if e, ok := err.(interface{ NoBreak() bool }); ok && !e.NoBreak() {
					break
				}
			}

			if i >= len(tc.results) {
				t.Errorf("parsed too many: %v", i)
				i++
				continue
			}

			if eq, ok := v.(interface{ Equals(interface{}) bool }); ok {
				if !eq.Equals(tc.results[i]) {
					t.Errorf("expected value %v, got %v", tc.results[i], v)
				}
			} else {
				if !reflect.DeepEqual(v, tc.results[i]) {
					t.Errorf("expected value %v, got %#v", tc.results[i], v)
				}
			}

			i++
		}
	}

	tcs := []testcase{
		{
			jsonString: `{
  "k":"v",
  "answer":42,
  "correct":true,
  "tags": [ "omg", 3.141 ]
}
{
  "k":"not v",
  "answer": 23,
  "correct":false,
  "tags": [ "conspiracy", 5 ]
}`,
			tipe: jsonType{},
			results: []interface{}{
				&jsonType{
					K:       "v",
					Answer:  42,
					Correct: true,
					Tags:    []interface{}{"omg", 3.141},
				},
				&jsonType{
					K:       "not v",
					Answer:  23,
					Correct: false,
					Tags:    []interface{}{"conspiracy", 5.0},
				},
			},
			errchk: func(err error) error {
				return err
			},
		},
	}

	for i, tc := range tcs {
		t.Logf("running test case %v", i)
		test(tc)
	}
}

type writeCloser struct {
	io.Writer
}

func (_ writeCloser) Close() error { return nil }

func TestSink(t *testing.T) {
	type jsonType struct {
		K       string        `json:"k"`
		Answer  int           `json:"answer"`
		Correct bool          `json:"correct"`
		Tags    []interface{} `json:"tags"`
	}

	type testcase struct {
		jsonString string
		values     []interface{}
		tipe       interface{}
		errchk     func(int, error) error
	}

	test := func(tc testcase) {
		ctx := context.Background()
		buf := bytes.NewBuffer([]byte(tc.jsonString))
		sink := NewSink(writeCloser{buf})

		dec := json.NewDecoder(buf)

		for i, v := range tc.values {
			err := sink.Pour(ctx, v)
			if err = tc.errchk(i, err); err != nil {
				t.Errorf("unexpected error %s in test case %#v\ni:%v", err, tc, i)
				if e, ok := err.(interface{ NoBreak() bool }); ok && !e.NoBreak() {
					break
				}
			}

			dst := reflect.New(reflect.TypeOf(tc.tipe)).Interface()
			if err = dec.Decode(dst); err != nil {
				t.Errorf("unexpected json decoding error %s in test case #%d", err, i)
			}

			if !reflect.DeepEqual(dst, v) {
				t.Errorf("expected %#v, got %#v", dst, v)
			}
		}
	}

	tcs := []testcase{
		{
			values: []interface{}{
				&jsonType{
					K:       "v",
					Answer:  42,
					Correct: true,
					Tags:    []interface{}{"omg", 3.141},
				},
				&jsonType{
					K:       "not v",
					Answer:  23,
					Correct: false,
					Tags:    []interface{}{"conspiracy", 5.0},
				},
			},
			tipe: jsonType{},
			errchk: func(i int, err error) error {
				return err
			},
		},
	}

	for i, tc := range tcs {
		t.Logf("running test case %v", i)
		test(tc)
	}
}

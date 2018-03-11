package json // import "cryptoscope.co/go/luigi/json"

import (
	"context"
	"encoding/json"
	"io"
	"reflect"

	"cryptoscope.co/go/luigi"
)

type source struct {
	t   reflect.Type
	dec *json.Decoder
}

// NewSource returns a new source that emits values of the pointer type as t read from the Reader in JSON format.
func NewSource(r io.Reader, t interface{}) luigi.Source {
	return &source{
		t:   reflect.TypeOf(t),
		dec: json.NewDecoder(r),
	}
}

func (src *source) Next(ctx context.Context) (interface{}, error) {
	x := reflect.New(src.t).Interface()
	err := src.dec.Decode(x)

	if err == io.EOF {
		return x, luigi.EOS{}
	}

	return x, err
}

type sink struct {
	io.Closer
	enc *json.Encoder
}

// NewSink returns a new sink that writes incoming data to the passed WriteCloser in JSON format
func NewSink(wc io.WriteCloser) luigi.Sink {
	return &sink{
		Closer: wc,
		enc:    json.NewEncoder(wc),
	}
}

func (sink *sink) Pour(ctx context.Context, v interface{}) error {
	return sink.enc.Encode(v)
}

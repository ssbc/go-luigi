package luigi // import "cryptoscope.co/go/luigi"

import (
	"context"
	"encoding/json"
	"io"
	"reflect"
)

type jsonSource struct {
	t   reflect.Type
	dec *json.Decoder
}

func NewJSONSource(r io.Reader, t interface{}) Source {
	return &jsonSource{
		t:   reflect.TypeOf(t),
		dec: json.NewDecoder(r),
	}
}

func (src *jsonSource) Next(ctx context.Context) (interface{}, error) {
	x := reflect.New(src.t).Interface()
	err := src.dec.Decode(x)

	if err == io.EOF {
		return x, EOS{}
	}

	return x, err
}

type jsonSink struct {
	io.Closer
	enc *json.Encoder
}

func NewJSONSink(wc io.WriteCloser) Sink {
	return &jsonSink{
		Closer: wc,
		enc:    json.NewEncoder(wc),
	}
}

func (sink *jsonSink) Pour(ctx context.Context, v interface{}) error {
	return sink.enc.Encode(v)
}

package luigi

import "context"

// SliceSink binds Source methods to an interface array.
type SliceSource []interface{}

// Next implements the Source interface.
func (src *SliceSource) Next(context.Context) (v interface{}, err error) {
	if len(*src) == 0 {
		return nil, EOS{}
	}

	v, *src = (*src)[0], (*src)[1:]

	return v, nil
}

// SliceSink binds Sink methods to an interface array.
type SliceSink []interface{}

// Pour implements the Sink interface.  It writes value to a destination Sink.
func (sink *SliceSink) Pour(ctx context.Context, v interface{}) error {
	*sink = append(*sink, v)
	return nil
}

// Close is a dummy method to implement the Sink interface.  Further writes
// this sink will succeed.
func (sink *SliceSink) Close() error { return nil }

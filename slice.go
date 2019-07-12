package luigi

import (
	"context"
	"errors"
)

type SliceSource []interface{}

func (src *SliceSource) Next(context.Context) (v interface{}, err error) {
	if len(*src) == 0 {
		return nil, EOS{}
	}

	v, *src = (*src)[0], (*src)[1:]

	return v, nil
}

// SliceSink binds Sink methods to an interface array.
type SliceSink struct {
	slice  *[]interface{}
	closed bool
}

// NewSliceSink returns a new SliceSink bound to the given interface array.
func NewSliceSink(arg *[]interface{}) *SliceSink {
	return &SliceSink{
		slice:  arg,
		closed: false,
	}
}

func (sink *SliceSink) Pour(ctx context.Context, v interface{}) error {
	if sink.closed {
		return errors.New("pour to closed sink")
	}
	*sink.slice = append(*sink.slice, v)
	return nil
}

func (sink *SliceSink) Close() error {
	sink.closed = true
	return nil
}

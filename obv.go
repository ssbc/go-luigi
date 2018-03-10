package luigi // import "cryptoscope.co/go/luigi"

import (
	"context"
	"sync"
)

// Observabe wraps an interface{} value and allows tracking changes to it
// TODO should Set and Value get a ctx?
// TODO should this really be an interface? Why? Why not?
type Observable interface {
	// Broadcast allows subscribing to changes
	Broadcast

	// Set sets a new value
	Set(interface{}) error

	// Value returns the current value
	Value() (interface{}, error)
}

// NewObservable returns a new Observable
func NewObservable(v interface{}) Observable {
	bcstSink, bcst := NewBroadcast()

	return &observable{
		Broadcast: bcst,
		sink:      bcstSink,
		v:         v,
	}
}

// observable is a concrete type implementing Observable
type observable struct {
	sync.Mutex
	Broadcast
	sink Sink

	v interface{}
}

// Set sets a new value
func (o *observable) Set(v interface{}) error {
	o.Lock()
	defer o.Unlock()

	o.v = v
	return o.sink.Pour(context.TODO(), v)
}

// Value returns the current value
func (o *observable) Value() (interface{}, error) {
	o.Lock()
	defer o.Unlock()

	return o.v, nil
}

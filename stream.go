package luigi // import "cryptoscope.co/go/luigi"

import "context"

type EOS struct{}

func (_ EOS) Error() string { return "end of stream" }
func IsEOS(err error) bool  { _, ok := err.(EOS); return ok }

type Sink interface {
	Pour(context.Context, interface{}) error
	Close() error
}

type Source interface {
	Next(context.Context) (interface{}, error)
}

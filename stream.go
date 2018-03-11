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

// Pump moves values from a source into a sink.
//
// Currently this doesn't work atomically, so if a Sink errors in the
// Pour call, the value that was read from the source is lost.
func Pump(ctx context.Context, sink Sink, src Source) error {
	for {
		v, err := src.Next(ctx)
		if IsEOS(err) {
			return nil
		} else if err != nil {
			return err
		}

		err = sink.Pour(ctx, v)
		if err != nil {
			return err
		}
	}

	panic("unreachable")
}

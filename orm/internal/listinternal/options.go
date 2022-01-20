package listinternal

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Options is the internal list options struct.
type Options struct {
	Start, End, Prefix []protoreflect.Value
	Reverse            bool
	Cursor             []byte
}

func (o Options) Validate() error {
	if o.Start != nil || o.End != nil {
		if o.Prefix != nil {
			return fmt.Errorf("can either use Start/End or Prefix, not both")
		}
	}

	return nil
}

type Option interface {
	apply(*Options)
}

type FuncOption func(*Options)

func (f FuncOption) apply(options *Options) {
	f(options)
}

func ApplyOptions(opts *Options, funcOpts []Option) {
	for _, opt := range funcOpts {
		opt.apply(opts)
	}
}

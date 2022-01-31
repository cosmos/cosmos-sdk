package listinternal

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Options is the internal list options struct.
type Options struct {
	Reverse       bool
	Cursor        []byte
	Filter        func(proto.Message) bool
	Offset, Limit uint64
	CountTotal    bool
}

func (o Options) Validate() error {
	if len(o.Cursor) != 0 {
		if o.Offset > 0 {
			return fmt.Errorf("can only specify one of cursor or offset")
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

package cachemulti

import (
	"io"

	"cosmossdk.io/store/cachemulti"
	"cosmossdk.io/store/types"
)

var NoopCloser io.Closer = CloserFunc(func() error { return nil })

type CloserFunc func() error

func (fn CloserFunc) Close() error {
	return fn()
}

// Store wraps sdk's cachemulti.Store to add io.Closer interface
type Store struct {
	cachemulti.Store
	io.Closer
}

func NewStore(
	stores map[types.StoreKey]types.CacheWrapper,
	traceWriter io.Writer, traceContext types.TraceContext,
	closer io.Closer,
) Store {
	if closer == nil {
		closer = NoopCloser
	}
	store := cachemulti.NewStore(nil, stores, nil, traceWriter, traceContext)
	return Store{
		Store:  store,
		Closer: closer,
	}
}

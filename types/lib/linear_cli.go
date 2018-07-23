package lib

import (
	"io"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

type linearCLIStore struct {
	ctx       context.CoreContext
	storeName string
	prefix    []byte
}

// Implements sdk.CacheWrapper
func (store linearCLIStore) CacheWrap() sdk.CacheWrap {
	panic("Should not reach here")
}

// Implements sdk.KVStore
func (store linearCLIStore) CacheWrapWithTrace(w io.Writer, tc sdk.TraceContext) sdk.CacheWrap {
	panic("Should not reach here")
}

// Implements sdk.Store
func (store linearCLIStore) GetStoreType() sdk.StoreType {
	panic("Should not reach here")
}

// Implements sdk.KVStore
func (store linearCLIStore) Get(key []byte) []byte {
	bz, err := store.ctx.QueryStore(append(store.prefix, key...), store.storeName)
	if err != nil {
		panic(err)
	}
	return bz
}

// Implements sdk.KVStore
func (store linearCLIStore) Has(key []byte) bool {
	panic("Should not reach here")
}

// Implements sdk.KVStore
func (store linearCLIStore) Set(key, value []byte) {
	panic("Should not reach here")
}

// Implements sdk.KVStore
func (store linearCLIStore) Delete(key []byte) {
	panic("Should not reach here")
}

// Implements sdk.KVstore
func (store linearCLIStore) Prefix(prefix []byte) sdk.KVStore {
	panic("Should not reach here")
}

// Implements sdk.KVStore
func (store linearCLIStore) Iterator(start, end []byte) sdk.Iterator {
	panic("Should not reach here")
}

// Implements sdk.KVStore
func (store linearCLIStore) ReverseIterator(start, end []byte) sdk.Iterator {
	panic("Should not reach here")
}

type linearClient struct {
	Linear
}

type LinearClient interface {
	Len() uint64
	Get(uint64, interface{}) error
	Iterate(interface{}, func(uint64) bool)

	Peek(interface{}) error
	IsEmpty() bool
}

func NewLinearClient(ctx context.CoreContext, storeName string, cdc *wire.Codec, prefix []byte, keys *LinearKeys) LinearClient {
	if keys == nil {
		keys = cachedDefaultLinearKeys
	}
	if keys.LengthKey == nil || keys.ElemKey == nil || keys.TopKey == nil {
		panic("Invalid LinearKeys")
	}
	return linearClient{
		Linear: Linear{
			cdc: cdc,
			store: linearCLIStore{
				ctx:       ctx,
				storeName: storeName,
				prefix:    prefix,
			},
			keys: keys,
		},
	}
}

func (cli linearClient) Iterate(ptr interface{}, fn func(uint64) bool) {
	top := cli.getTop()
	length := cli.Len()

	var i uint64
	for i = top; i < length; i++ {
		if cli.Get(i, ptr) != nil {
			continue
		}
		if fn(i) {
			break
		}
	}
}

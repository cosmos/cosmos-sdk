package commitment

import (
	"io"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/types"
)

var _ types.KVStore = Store{}

// Store is a simple []byte to []byte map
// panics when there is no corresponding value(to be compatible with KVSTore interface)
// Set/Delete/Iterator/ReverseIterator method should not be used
type Store struct {
	m map[string][]byte
}

func NewStore(root Root, proofs []FullProof) (store Store, err error) {
	for _, proof := range proofs {
		err = proof.Verify(root)
		if err != nil {
			return
		}
	}

	store.m = make(map[string][]byte)

	for _, proof := range proofs {
		store.m[string(proof.Key)] = proof.Value
	}

	return
}

func (store Store) GetStoreType() types.StoreType {
	return types.StoreTypeTransient // XXX: check is it right
}

func (store Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(store)
}

func (store Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	// FIXME
	return store.CacheWrap()
}

func (store Store) Get(key []byte) []byte {
	res, ok := store.m[string(key)]
	if !ok {
		panic(NoProofError{})
	}
	return res
}

func (store Store) Set(key, value []byte) {
	panic(MutableMethodError{})
}

// TODO: consider using this method to check whether the proof provided or not
// which may violate KVStore semantics
func (store Store) Has(key []byte) bool {
	return store.Get(key) != nil
}

func (store Store) Delete(key []byte) {
	panic(MutableMethodError{})
}

func (store Store) Iterator(begin, end []byte) types.Iterator {
	panic(IteratorMethodError{})
}

func (store Store) ReverseIterator(begin, end []byte) types.Iterator {
	panic(IteratorMethodError{})
}

type NoProofError struct {
	// XXX
}

type MutableMethodError struct {
	// XXX
}

type IteratorMethodError struct {
}

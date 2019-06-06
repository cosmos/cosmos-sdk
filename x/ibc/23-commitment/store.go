package commitment

import (
	"io"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/types"
)

var _ types.KVStore = Store{}

// Panics when there is no corresponding proof or the proof is invalid
// (to be compatible with KVStore interface)
// The semantics of the methods are redefined and does not compatible(should be improved)
// Set -> Proof corresponding to the provided key is verified with the provided value
// Has -> Returns true if the proof is verified, returns false in any other case
// Other methods should not be used
type Store struct {
	root     Root
	proofs   map[string]Proof
	verified map[string]struct{}
}

// Proofs must be provided
func NewStore(root Root, proofs []Proof) Store {
	return Store{
		root:     root,
		proofs:   proofs,
		verified: make(map[string]struct{}),
	}
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
	panic(MethodError{})
}

func (store Store) Set(key, value []byte) {
	proof, ok := store.proofs[string(key)]
	if !ok {
		return
	}
	err := proof.Verify(store.root, key, value)
	if err == nil {
		store.verified[string(key)] = struct{}{}
	}

	return
}

// TODO: consider using this method to check whether the proof provided or not
// which may violate KVStore semantics
func (store Store) Has(key []byte) bool {
	_, ok := store.verified[string(key)]
	return ok
}

func (store Store) Delete(key []byte) {
	panic(MethodError{})
}

func (store Store) Iterator(begin, end []byte) types.Iterator {
	panic(MethodError{})
}

func (store Store) ReverseIterator(begin, end []byte) types.Iterator {
	panic(MethodError{})
}

type NoProofError struct {
	// XXX
}

type MethodError struct {
	// XXX
}

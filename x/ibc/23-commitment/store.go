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

// OperationType might be useful in future but not currently
/*
type OperationType byte

const (
	Get OperationType = iota
	Set
	Has
	Delete
)
*/

type Operation struct {
	// Type OperationType
	Key []byte
}

var _ types.KVStore = (*KeyLoggerStore)(nil)

// TraceStore is a dummy KVStore which logs all method call on it
// but returns empty vaule all times
type KeyLoggerStore struct {
	ops []Operation
}

func (*KeyLoggerStore) GetStoreType() types.StoreType {
	return types.StoreTypeTransient
}

func (store *KeyLoggerStore) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(store)
}

func (store *KeyLoggerStore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	// FIXME
	return store.CacheWrap()
}

func (store *KeyLoggerStore) log(key []byte) {
	store.ops = append(store.ops, Operation{key})
}

func (store *KeyLoggerStore) Get(key []byte) []byte {
	store.log(key)
	return nil
}

func (store *KeyLoggerStore) Set(key, value []byte) {
	store.log(key)
}

func (store *KeyLoggerStore) Has(key []byte) bool {
	store.log(key)
	return false
}

func (store *KeyLoggerStore) Delete(key []byte) {
	store.log(key)
}

func (store *KeyLoggerStore) Iterator(begin, end []byte) types.Iterator {
	// XXX: should return empty iterator instead of nil
	return nil
}

func (store *KeyLoggerStore) ReverseIterator(begin, end []byte) types.Iterator {
	return nil
}

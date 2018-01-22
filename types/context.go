package types

import (
	"context"
	"sync"

	"github.com/golang/protobuf/proto"

	abci "github.com/tendermint/abci/types"
)

// TODO: Add a default logger.

/*
The intent of Context is for it to be an immutable object that can be
cloned and updated cheaply with WithValue() and passed forward to the
next decorator or handler. For example,

 func MsgHandler(ctx Context, tx Tx) Result {
 	...
 	ctx = ctx.WithValue(key, value)
 	...
 }
*/
type Context struct {
	context.Context
	pst *thePast
	gen int
	// Don't add any other fields here,
	// it's probably not what you want to do.
}

func NewContext(ms MultiStore, header abci.Header, isCheckTx bool, txBytes []byte) Context {
	c := Context{
		Context: context.Background(),
		pst:     newThePast(),
		gen:     0,
	}
	c = c.withMultiStore(ms)
	c = c.withBlockHeader(header)
	c = c.withBlockHeight(header.Height)
	c = c.withChainID(header.ChainID)
	c = c.withIsCheckTx(isCheckTx)
	c = c.withTxBytes(txBytes)
	return c
}

//----------------------------------------
// Getting a value

func (c Context) Value(key interface{}) interface{} {
	value := c.Context.Value(key)
	if cloner, ok := value.(cloner); ok {
		return cloner.Clone()
	}
	if message, ok := value.(proto.Message); ok {
		return proto.Clone(message)
	}
	return value
}

// KVStore fetches a KVStore from the MultiStore.
func (c Context) KVStore(key StoreKey) KVStore {
	return c.multiStore().GetKVStore(key)
}

//----------------------------------------
// With* (setting a value)

func (c Context) WithValue(key interface{}, value interface{}) Context {
	return c.withValue(key, value)
}

func (c Context) WithCloner(key interface{}, value cloner) Context {
	return c.withValue(key, value)
}

func (c Context) WithCacheWrapper(key interface{}, value CacheWrapper) Context {
	return c.withValue(key, value)
}

func (c Context) WithProtoMsg(key interface{}, value proto.Message) Context {
	return c.withValue(key, value)
}

func (c Context) WithMultiStore(key StoreKey, ms MultiStore) Context {
	return c.withValue(key, ms)
}

func (c Context) WithString(key interface{}, value string) Context {
	return c.withValue(key, value)
}

func (c Context) WithInt32(key interface{}, value int32) Context {
	return c.withValue(key, value)
}

func (c Context) WithUint32(key interface{}, value uint32) Context {
	return c.withValue(key, value)
}

func (c Context) WithUint64(key interface{}, value uint64) Context {
	return c.withValue(key, value)
}

func (c Context) withValue(key interface{}, value interface{}) Context {
	c.pst.bump(Op{
		gen:   c.gen + 1,
		key:   key,
		value: value,
	}) // increment version for all relatives.

	return Context{
		Context: context.WithValue(c.Context, key, value),
		pst:     c.pst,
		gen:     c.gen + 1,
	}
}

//----------------------------------------
// Values that require no key.

type contextKey int // local to the context module

const (
	contextKeyMultiStore contextKey = iota
	contextKeyBlockHeader
	contextKeyBlockHeight
	contextKeyChainID
	contextKeyIsCheckTx
	contextKeyTxBytes
)

// NOTE: Do not expose MultiStore, to require the store key.
func (c Context) multiStore() MultiStore {
	return c.Value(contextKeyMultiStore).(MultiStore)
}

func (c Context) BlockHeader() abci.Header {
	return c.Value(contextKeyBlockHeader).(abci.Header)
}

func (c Context) BlockHeight() int64 {
	return c.Value(contextKeyBlockHeight).(int64)
}

func (c Context) ChainID() string {
	return c.Value(contextKeyChainID).(string)
}

func (c Context) IsCheckTx() bool {
	return c.Value(contextKeyIsCheckTx).(bool)
}

func (c Context) TxBytes() []byte {
	return c.Value(contextKeyTxBytes).([]byte)
}

// Unexposed to prevent overriding.
func (c Context) withMultiStore(ms MultiStore) Context {
	return c.withValue(contextKeyMultiStore, ms)
}

// Unexposed to prevent overriding.
func (c Context) withBlockHeader(header abci.Header) Context {
	var _ proto.Message = &header // for cloning.
	return c.withValue(contextKeyBlockHeader, header)
}

// Unexposed to prevent overriding.
func (c Context) withBlockHeight(height int64) Context {
	return c.withValue(contextKeyBlockHeight, height)
}

// Unexposed to prevent overriding.
func (c Context) withChainID(chainID string) Context {
	return c.withValue(contextKeyChainID, chainID)
}

// Unexposed to prevent overriding.
func (c Context) withIsCheckTx(isCheckTx bool) Context {
	return c.withValue(contextKeyIsCheckTx, isCheckTx)
}

// Unexposed to prevent overriding.
func (c Context) withTxBytes(txBytes []byte) Context {
	return c.withValue(contextKeyTxBytes, txBytes)
}

//----------------------------------------
// thePast

// Returns false if ver > 0.
// The first operation is version 1.
func (c Context) GetOp(ver int64) (Op, bool) {
	return c.pst.getOp(ver)
}

//----------------------------------------
// Misc.

type cloner interface {
	Clone() interface{} // deep copy
}

type Op struct {
	// type is always 'with'
	gen   int
	key   interface{}
	value interface{}
}

type thePast struct {
	mtx sync.RWMutex
	ver int
	ops []Op
}

func newThePast() *thePast {
	return &thePast{
		ver: 0,
		ops: nil,
	}
}

func (pst *thePast) bump(op Op) {
	pst.mtx.Lock()
	pst.ver += 1
	pst.ops = append(pst.ops, op)
	pst.mtx.Unlock()
}

func (pst *thePast) version() int {
	pst.mtx.RLock()
	defer pst.mtx.RUnlock()
	return pst.ver
}

// Returns false if ver > 0.
// The first operation is version 1.
func (pst *thePast) getOp(ver int64) (Op, bool) {
	pst.mtx.RLock()
	defer pst.mtx.RUnlock()
	l := int64(len(pst.ops))
	if l < ver {
		return Op{}, false
	} else {
		return pst.ops[ver-1], true
	}
}

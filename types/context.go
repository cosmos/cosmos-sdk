// nolint
package types

import (
	"context"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/gaskv"
	stypes "github.com/cosmos/cosmos-sdk/store/types"
)

/*
Context is an immutable object contains all information needed to
process a request.
*/
type Context struct {
	ctx           context.Context
	ms            MultiStore
	header        abci.Header
	chainID       string
	txBytes       []byte
	logger        log.Logger
	voteInfo      []abci.VoteInfo
	gasMeter      GasMeter
	blockGasMeter GasMeter
	checkTx       bool
	minGasPrice   DecCoins
	consParams    *abci.ConsensusParams
	evtManager    *EventManager
}

// Proposed rename, not done to avoid API breakage
type Request = Context

// Read-only accessors
func (c Context) Context() context.Context    { return c.ctx }
func (c Context) MultiStore() MultiStore      { return c.ms }
func (c Context) BlockHeight() int64          { return c.header.Height }
func (c Context) BlockTime() time.Time        { return c.header.Time }
func (c Context) ChainID() string             { return c.chainID }
func (c Context) TxBytes() []byte             { return c.txBytes }
func (c Context) Logger() log.Logger          { return c.logger }
func (c Context) VoteInfos() []abci.VoteInfo  { return c.voteInfo }
func (c Context) GasMeter() GasMeter          { return c.gasMeter }
func (c Context) BlockGasMeter() GasMeter     { return c.blockGasMeter }
func (c Context) IsCheckTx() bool             { return c.checkTx }
func (c Context) MinGasPrices() DecCoins      { return c.minGasPrice }
func (c Context) EventManager() *EventManager { return c.evtManager }

// clone the header before returning
func (c Context) BlockHeader() abci.Header {
	// TODO: figure out clone better
	return c.header
	// var msg = proto.Clone(&c.header).(*abci.Header)
	// return *msg
}

func (c Context) ConsensusParams() *abci.ConsensusParams {
	return c.consParams
	// return proto.Clone(c.consParams).(*abci.ConsensusParams)
}

// create a new context
func NewContext(ms MultiStore, header abci.Header, isCheckTx bool, logger log.Logger) Context {
	return Context{
		ctx:         context.Background(),
		ms:          ms,
		header:      header,
		chainID:     header.ChainID,
		checkTx:     isCheckTx,
		logger:      logger,
		gasMeter:    stypes.NewInfiniteGasMeter(),
		minGasPrice: DecCoins{},
		evtManager:  NewEventManager(),
	}
}

func (c Context) WithContext(ctx context.Context) Context {
	c.ctx = ctx
	return c
}

func (c Context) WithMultiStore(ms MultiStore) Context {
	c.ms = ms
	return c
}

func (c Context) WithBlockHeader(header abci.Header) Context {
	c.header = header
	return c
}

func (c Context) WithBlockTime(newTime time.Time) Context {
	newHeader := c.BlockHeader()
	newHeader.Time = newTime
	return c.WithBlockHeader(newHeader)
}

func (c Context) WithProposer(addr ConsAddress) Context {
	newHeader := c.BlockHeader()
	newHeader.ProposerAddress = addr.Bytes()
	return c.WithBlockHeader(newHeader)
}

func (c Context) WithBlockHeight(height int64) Context {
	newHeader := c.BlockHeader()
	newHeader.Height = height
	return c.WithBlockHeader(newHeader)
}

func (c Context) WithChainID(chainID string) Context {
	c.chainID = chainID
	return c
}

func (c Context) WithTxBytes(txBytes []byte) Context {
	c.txBytes = txBytes
	return c
}

func (c Context) WithLogger(logger log.Logger) Context {
	c.logger = logger
	return c
}

func (c Context) WithVoteInfos(voteInfo []abci.VoteInfo) Context {
	c.voteInfo = voteInfo
	return c
}

func (c Context) WithGasMeter(meter GasMeter) Context {
	c.gasMeter = meter
	return c
}

func (c Context) WithBlockGasMeter(meter GasMeter) Context {
	c.blockGasMeter = meter
	return c
}

func (c Context) WithIsCheckTx(isCheckTx bool) Context {
	c.checkTx = isCheckTx
	return c
}

func (c Context) WithMinGasPrices(gasPrices DecCoins) Context {
	c.minGasPrice = gasPrices
	return c
}

func (c Context) WithConsensusParams(params *abci.ConsensusParams) Context {
	c.consParams = params
	return c
}

func (c Context) WithEventManager(em *EventManager) Context {
	c.evtManager = em
	return c
}

// TODO: remove???
func (c Context) IsZero() bool {
	return c.ms == nil
}

// // context value for the provided key
// func (c Context) Value(key interface{}) interface{} {
// 	value := c.Context.Value(key)
// 	if cloner, ok := value.(cloner); ok {
// 		return cloner.Clone()
// 	}
// 	if message, ok := value.(proto.Message); ok {
// 		return proto.Clone(message)
// 	}
// 	return value
// }

// ----------------------------------------------------------------------------
// Setters
// ----------------------------------------------------------------------------

// func (c Context) WithValue(key interface{}, value interface{}) Context {
// 	return c.withValue(key, value)
// }
// func (c Context) WithCloner(key interface{}, value cloner) Context {
// 	return c.withValue(key, value)
// }
// func (c Context) WithCacheWrapper(key interface{}, value CacheWrapper) Context {
// 	return c.withValue(key, value)
// }
// func (c Context) WithProtoMsg(key interface{}, value proto.Message) Context {
// 	return c.withValue(key, value)
// }
// func (c Context) WithString(key interface{}, value string) Context {
// 	return c.withValue(key, value)
// }
// func (c Context) WithInt32(key interface{}, value int32) Context {
// 	return c.withValue(key, value)
// }
// func (c Context) WithUint32(key interface{}, value uint32) Context {
// 	return c.withValue(key, value)
// }
// func (c Context) WithUint64(key interface{}, value uint64) Context {
// 	return c.withValue(key, value)
// }

// func (c Context) withValue(key interface{}, value interface{}) Context {
// 	c.pst.bump(Op{
// 		gen:   c.gen + 1,
// 		key:   key,
// 		value: value,
// 	}) // increment version for all relatives.

// 	return Context{
// 		Context: context.WithValue(c.Context, key, value),
// 		pst:     c.pst,
// 		gen:     c.gen + 1,
// 	}
// }

// ----------------------------------------------------------------------------
// Store / Caching
// ----------------------------------------------------------------------------

// KVStore fetches a KVStore from the MultiStore.
func (c Context) KVStore(key StoreKey) KVStore {
	return gaskv.NewStore(c.MultiStore().GetKVStore(key), c.GasMeter(), stypes.KVGasConfig())
}

// TransientStore fetches a TransientStore from the MultiStore.
func (c Context) TransientStore(key StoreKey) KVStore {
	return gaskv.NewStore(c.MultiStore().GetKVStore(key), c.GasMeter(), stypes.TransientGasConfig())
}

// Cache the multistore and return a new cached context. The cached context is
// written to the context when writeCache is called.
func (c Context) CacheContext() (cc Context, writeCache func()) {
	cms := c.MultiStore().CacheMultiStore()
	cc = c.WithMultiStore(cms)
	return cc, cms.Write
}

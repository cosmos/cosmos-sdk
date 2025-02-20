package types

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/store/gaskv"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

/*
Context is an immutable object contains all information needed to
process a request.

It contains a context.Context object inside if you want to use that,
but please do not over-use it. We try to keep all data structured
and standard additions here would be better just to add to the Context struct
*/
type Context struct {
	baseCtx              context.Context
	ms                   MultiStore
	header               tmproto.Header
	headerHash           tmbytes.HexBytes
	chainID              string
	txBytes              []byte
	logger               log.Logger
	voteInfo             []abci.VoteInfo
	gasMeter             GasMeter
	blockGasMeter        GasMeter
	checkTx              bool
	recheckTx            bool // if recheckTx == true, then checkTx must also be true
	minGasPrice          DecCoins
	consParams           *tmproto.ConsensusParams
	eventManager         *EventManager
	priority             int64 // The tx priority, only relevant in CheckTx
	kvGasConfig          storetypes.GasConfig
	transientKVGasConfig storetypes.GasConfig
}

// Proposed rename, not done to avoid API breakage
type Request = Context

// Read-only accessors
func (c Context) Context() context.Context                   { return c.baseCtx }
func (c Context) MultiStore() MultiStore                     { return c.ms }
func (c Context) BlockHeight() int64                         { return c.header.Height }
func (c Context) BlockTime() time.Time                       { return c.header.Time }
func (c Context) ChainID() string                            { return c.chainID }
func (c Context) TxBytes() []byte                            { return c.txBytes }
func (c Context) Logger() log.Logger                         { return c.logger }
func (c Context) VoteInfos() []abci.VoteInfo                 { return c.voteInfo }
func (c Context) GasMeter() GasMeter                         { return c.gasMeter }
func (c Context) BlockGasMeter() GasMeter                    { return c.blockGasMeter }
func (c Context) IsCheckTx() bool                            { return c.checkTx }
func (c Context) IsReCheckTx() bool                          { return c.recheckTx }
func (c Context) MinGasPrices() DecCoins                     { return c.minGasPrice }
func (c Context) EventManager() *EventManager                { return c.eventManager }
func (c Context) Priority() int64                            { return c.priority }
func (c Context) KVGasConfig() storetypes.GasConfig          { return c.kvGasConfig }
func (c Context) TransientKVGasConfig() storetypes.GasConfig { return c.transientKVGasConfig }

// clone the header before returning
func (c Context) BlockHeader() tmproto.Header {
	msg := proto.Clone(&c.header).(*tmproto.Header)
	return *msg
}

// HeaderHash returns a copy of the header hash obtained during abci.RequestBeginBlock
func (c Context) HeaderHash() tmbytes.HexBytes {
	hash := make([]byte, len(c.headerHash))
	copy(hash, c.headerHash)
	return hash
}

func (c Context) ConsensusParams() *tmproto.ConsensusParams {
	return proto.Clone(c.consParams).(*tmproto.ConsensusParams)
}

func (c Context) Deadline() (deadline time.Time, ok bool) {
	return c.baseCtx.Deadline()
}

func (c Context) Done() <-chan struct{} {
	return c.baseCtx.Done()
}

func (c Context) Err() error {
	return c.baseCtx.Err()
}

// create a new context
func NewContext(ms MultiStore, header tmproto.Header, isCheckTx bool, logger log.Logger) Context {
	// https://github.com/gogo/protobuf/issues/519
	header.Time = header.Time.UTC()
	return Context{
		baseCtx:              context.Background(),
		ms:                   ms,
		header:               header,
		chainID:              header.ChainID,
		checkTx:              isCheckTx,
		logger:               logger,
		gasMeter:             storetypes.NewInfiniteGasMeter(),
		minGasPrice:          DecCoins{},
		eventManager:         NewEventManager(),
		kvGasConfig:          storetypes.KVGasConfig(),
		transientKVGasConfig: storetypes.TransientGasConfig(),
	}
}

// WithContext returns a Context with an updated context.Context.
func (c Context) WithContext(ctx context.Context) Context {
	c.baseCtx = ctx
	return c
}

// WithMultiStore returns a Context with an updated MultiStore.
func (c Context) WithMultiStore(ms MultiStore) Context {
	c.ms = ms
	return c
}

// WithBlockHeader returns a Context with an updated tendermint block header in UTC time.
func (c Context) WithBlockHeader(header tmproto.Header) Context {
	// https://github.com/gogo/protobuf/issues/519
	header.Time = header.Time.UTC()
	c.header = header
	return c
}

// WithHeaderHash returns a Context with an updated tendermint block header hash.
func (c Context) WithHeaderHash(hash []byte) Context {
	temp := make([]byte, len(hash))
	copy(temp, hash)

	c.headerHash = temp
	return c
}

// WithBlockTime returns a Context with an updated tendermint block header time in UTC with no monotonic component.
// Stripping the monotonic component is for time equality.
func (c Context) WithBlockTime(newTime time.Time) Context {
	newHeader := c.BlockHeader()
	// https://github.com/gogo/protobuf/issues/519
	newHeader.Time = newTime.Round(0).UTC()
	return c.WithBlockHeader(newHeader)
}

// WithProposer returns a Context with an updated proposer consensus address.
func (c Context) WithProposer(addr ConsAddress) Context {
	newHeader := c.BlockHeader()
	newHeader.ProposerAddress = addr.Bytes()
	return c.WithBlockHeader(newHeader)
}

// WithBlockHeight returns a Context with an updated block height.
func (c Context) WithBlockHeight(height int64) Context {
	newHeader := c.BlockHeader()
	newHeader.Height = height
	return c.WithBlockHeader(newHeader)
}

// WithChainID returns a Context with an updated chain identifier.
func (c Context) WithChainID(chainID string) Context {
	c.chainID = chainID
	return c
}

// WithTxBytes returns a Context with an updated txBytes.
func (c Context) WithTxBytes(txBytes []byte) Context {
	c.txBytes = txBytes
	return c
}

// WithLogger returns a Context with an updated logger.
func (c Context) WithLogger(logger log.Logger) Context {
	c.logger = logger
	return c
}

// WithVoteInfos returns a Context with an updated consensus VoteInfo.
func (c Context) WithVoteInfos(voteInfo []abci.VoteInfo) Context {
	c.voteInfo = voteInfo
	return c
}

// WithGasMeter returns a Context with an updated transaction GasMeter.
func (c Context) WithGasMeter(meter GasMeter) Context {
	c.gasMeter = meter
	return c
}

// WithBlockGasMeter returns a Context with an updated block GasMeter
func (c Context) WithBlockGasMeter(meter GasMeter) Context {
	c.blockGasMeter = meter
	return c
}

// WithKVGasConfig returns a Context with an updated gas configuration for
// the KVStore
func (c Context) WithKVGasConfig(gasConfig storetypes.GasConfig) Context {
	c.kvGasConfig = gasConfig
	return c
}

// WithTransientKVGasConfig returns a Context with an updated gas configuration for
// the transient KVStore
func (c Context) WithTransientKVGasConfig(gasConfig storetypes.GasConfig) Context {
	c.transientKVGasConfig = gasConfig
	return c
}

// WithIsCheckTx enables or disables CheckTx value for verifying transactions and returns an updated Context
func (c Context) WithIsCheckTx(isCheckTx bool) Context {
	c.checkTx = isCheckTx
	return c
}

// WithIsRecheckTx called with true will also set true on checkTx in order to
// enforce the invariant that if recheckTx = true then checkTx = true as well.
func (c Context) WithIsReCheckTx(isRecheckTx bool) Context {
	if isRecheckTx {
		c.checkTx = true
	}
	c.recheckTx = isRecheckTx
	return c
}

// WithMinGasPrices returns a Context with an updated minimum gas price value
func (c Context) WithMinGasPrices(gasPrices DecCoins) Context {
	c.minGasPrice = gasPrices
	return c
}

// WithConsensusParams returns a Context with an updated consensus params
func (c Context) WithConsensusParams(params *tmproto.ConsensusParams) Context {
	c.consParams = params
	return c
}

// WithEventManager returns a Context with an updated event manager
func (c Context) WithEventManager(em *EventManager) Context {
	c.eventManager = em
	return c
}

// WithPriority returns a Context with an updated tx priority
func (c Context) WithPriority(p int64) Context {
	c.priority = p
	return c
}

// TODO: remove???
func (c Context) IsZero() bool {
	return c.ms == nil
}

func (c Context) WithValue(key, value interface{}) Context {
	c.baseCtx = context.WithValue(c.baseCtx, key, value)
	return c
}

func (c Context) Value(key interface{}) interface{} {
	if key == SdkContextKey {
		return c
	}

	return c.baseCtx.Value(key)
}

// ----------------------------------------------------------------------------
// Store / Caching
// ----------------------------------------------------------------------------

// KVStore fetches a KVStore from the MultiStore.
func (c Context) KVStore(key storetypes.StoreKey) KVStore {
	return gaskv.NewStore(c.MultiStore().GetKVStore(key), c.GasMeter(), c.kvGasConfig)
}

// TransientStore fetches a TransientStore from the MultiStore.
func (c Context) TransientStore(key storetypes.StoreKey) KVStore {
	return gaskv.NewStore(c.MultiStore().GetKVStore(key), c.GasMeter(), c.transientKVGasConfig)
}

// CacheContext returns a new Context with the multi-store cached and a new
// EventManager. The cached context is written to the context when writeCache
// is called. Note, events are automatically emitted on the parent context's
// EventManager when the caller executes the write.
func (c Context) CacheContext() (cc Context, writeCache func()) {
	cms := c.MultiStore().CacheMultiStore()
	cc = c.WithMultiStore(cms).WithEventManager(NewEventManager())

	writeCache = func() {
		c.EventManager().EmitEvents(cc.EventManager().Events())
		cms.Write()
	}

	return cc, writeCache
}

var _ context.Context = Context{}

// ContextKey defines a type alias for a stdlib Context key.
type ContextKey string

// SdkContextKey is the key in the context.Context which holds the sdk.Context.
const SdkContextKey ContextKey = "sdk-context"

// WrapSDKContext returns a stdlib context.Context with the provided sdk.Context's internal
// context as a value. It is useful for passing an sdk.Context  through methods that take a
// stdlib context.Context parameter such as generated gRPC methods. To get the original
// sdk.Context back, call UnwrapSDKContext.
func WrapSDKContext(ctx Context) context.Context {
	return ctx
}

// UnwrapSDKContext retrieves a Context from a context.Context instance
// attached with WrapSDKContext. It panics if a Context was not properly
// attached
func UnwrapSDKContext(ctx context.Context) Context {
	if sdkCtx, ok := ctx.(Context); ok {
		return sdkCtx
	}
	return ctx.Value(SdkContextKey).(Context)
}

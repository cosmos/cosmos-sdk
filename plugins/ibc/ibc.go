package ibc

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	tm "github.com/tendermint/tendermint/types"
)

const (
	// Key parts
	_IBC        = "ibc"
	_BLOCKCHAIN = "blockchain"
	_GENESIS    = "genesis"
	_STATE      = "state"
	_HASHES     = "hashes"
	_EGRESS     = "egress"
	_CONNECTION = "connection"
)

type IBCPluginState struct {
	// @[:ibc, :blockchain, :genesis, ChainID] <~ BlockchainGenesis
	// @[:ibc, :blockchain, :state, ChainID] <~ BlockchainState
	// @[:ibc, :blockchain, :hashes, ChainID, Height] <~ some blockhash []byte
	// @[:ibc, :egress, Src, Dst, Sequence] <~ Packet
	// @[:ibc, :connection, Src, Dst] <~ Connection # TODO - keep connection state
}

type BlockchainGenesis struct {
	ChainID string
	Genesis string
}

type BlockchainState struct {
	ChainID         string
	Validators      []tm.Validator
	LastBlockHash   []byte
	LastBlockHeight uint64
}

type Packet struct {
	SrcChainID string
	DstChainID string
	Sequence   uint64
	Type       string
	Payload    []byte
}

//--------------------------------------------------------------------------------

const (
	IBCTxTypeRegisterChain = byte(0x01)
	IBCTxTypeUpdateChain   = byte(0x02)
	IBCTxTypePacket        = byte(0x03)
)

var _ = wire.RegisterInterface(
	struct{ IBCTx }{},
	wire.ConcreteType{IBCRegisterChainTx{}, IBCTxTypeRegisterChain},
	wire.ConcreteType{IBCUpdateChainTx{}, IBCTxTypeUpdateChain},
	wire.ConcreteType{IBCPacketTx{}, IBCTxTypePacket},
)

type IBCTx interface {
	AssertIsIBCTx()
	ValidateBasic() abci.Result
}

func (IBCRegisterChainTx) AssertIsIBCTx() {}
func (IBCUpdateChainTx) AssertIsIBCTx()   {}
func (IBCPacketTx) AssertIsIBCTx()        {}

type IBCRegisterChainTx struct {
	BlockchainGenesis
}

func (IBCRegisterChainTx) ValidateBasic() abci.Result {
	// TODO - validate
	return
}

type IBCUpdateChainTx struct {
	Header tm.Header
	Commit tm.Commit
	// TODO: NextValidators
}

func (IBCUpdateChainTx) ValidateBasic() abci.Result {
	// TODO - validate
	return
}

type IBCPacketTx struct {
	FromChainID     string // The immediate source of the packet, not always Packet.SrcChainID
	FromChainHeight uint64 // The block height in which Packet was committed, to check Proof
	Packet
	Proof merkle.IAVLProof
}

func (IBCPacketTx) ValidateBasic() abci.Result {
	// TODO - validate
	return
}

//--------------------------------------------------------------------------------

type IBCPlugin struct {
}

func (ibc *IBCPlugin) Name() string {
	"IBC"
}

func (ibc *IBCPlugin) StateKey() []byte {
	return []byte(fmt.Sprintf("IBCPlugin.State", ibc.name))
}

func New(name string) *IBCPlugin {
	return &IBCPlugin{
		name: name,
	}
}

func (ibc *IBCPlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return ""
}

func (ibc *IBCPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
	// Decode tx
	var tx IBCTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Validate tx
	res := tx.ValidateBasic()
	if res.IsErr() {
		return res.PrependLog("ValidateBasic Failed: ")
	}

	// TODO - Check whether sufficient funds

	defer func() {
		// TODO - Refund any remaining funds left over
		// e.g. !ctx.Coins.Minus(tx.Fee).IsZero()
		// ctx.CallerAccount is synced w/ store, so just modify that and store it.
		// NOTE: We should use the CallContext to store fund/refund information.
	}()

	sm := &IBCStateMachine{store, ctx, abci.OK}

	switch tx := tx.(type) {
	case IBCRegisterChainTx:
		sm.runRegisterChainTx(tx)
	case IBCUpdateChainTx:
		sm.runUpdateChainTx(tx)
	case IBCPacketTx:
		sm.runPacketTx(tx)
	}

	return sm.res
}

type IBCStateMachine struct {
	store types.KVStore
	ctx   types.CallContext
	res   abci.Result
}

func (sm *IBCStateMachine) runRegisterChainTx(tx IBCRegisterChainTx) {
	chainGenKey := toKey(_IBC, _BLOCKCHAIN, _GENESIS, chain.ChainID)
	chainGen := tx.BlockchainGenesis

	// Make sure chainGen doesn't already exist
	if exists(sm.store, chainGenKey) {
		return // Already exists, do nothing
	}

	// Save new BlockchainGenesis
	save(sm.store, chainGenKey, chainGen)

	// TODO - Create and save new BlockchainState
}

func (sm *IBCStateMachine) runUpdateChainTx(tx IBCUpdateChainTx) {
	chainID := tx.Header.ChainID
	chainStateKey := toKey(_IBC, _BLOCKCHAIN, _STATE, chainID)

	// Make sure chainState exists
	if !exists(sm.store, chainStateKey) {
		return // Chain does not exist, do nothing
	}

	// Load latest chainState
	var chainState BlockchainState
	exists, err := load(sm.store, chainStateKey, &chainState)
	if err != nil {
		sm.res = abci.ErrInternalError.AppendLog("Loading ChainState: %v", err.Error())
		return
	}

	// TODO Compute blockHash from Header
	// TODO Check commit against validators
	// NOTE: Commit's votes include ValidatorAddress, so can be matched up against chainState.Validators
	//       for the demo we could assume that the validator set hadn't changed,
	//       though we should check that explicitly.
	// TODO Store blockhash
	// TODO Update chainState
	// TODO Store chainState
}

func (sm *IBCStateMachine) runPacketTx(tx IBCPacketTx) {
	// TODO Make sure packat doesn't already exist
	// TODO Load associated blockHash and make sure it exists
	// TODO compute packet key
	// TODO Make sure packet's proof matches given (packet, key, blockhash)
	// TODO Store packet
}

func (ibc *IBCPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
}

func (ibc *IBCPlugin) BeginBlock(store types.KVStore, height uint64) {
}

func (ibc *IBCPlugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	return nil
}

//--------------------------------------------------------------------------------
// TODO: move to utils

// Returns true if exists, false if nil.
func exists(store types.KVStore, key []byte) (exists bool) {
	value := store.Get(key)
	return len(value) > 0
}

// Load bytes from store by reading value for key and read into ptr.
// Returns true if exists, false if nil.
// Returns err if decoding error.
func load(store types.KVStore, key []byte, ptr interface{}) (exists bool, err error) {
	value := store.Get(key)
	if len(value) > 0 {
		err = wire.ReadBinaryBytes(value, ptr)
		if err != nil {
			return true, errors.New(
				Fmt("Error decoding key 0x%X = 0x%X: %v", key, value, err.Error()),
			)
		}
		return true, nil
	} else {
		return false, nil
	}
}

// Save bytes to store by writing obj's go-wire binary bytes.
func save(store types.KVStore, key []byte, obj interface{}) {
	store.Set(key, wire.BinaryBytes(obj))
}

// Key parts are URL escaped and joined with ','
func toKey(parts ...string) []byte {
	escParts := make([]string, len(parts))
	for i, part := range parts {
		escParts[i] = url.QueryEscape(part)
	}
	return []byte(strings.Join(escParts, ","))
}

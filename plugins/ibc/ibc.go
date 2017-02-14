package ibc

import (
	"bytes"
	"errors"
	"net/url"
	"strings"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
	merkle "github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
	tm "github.com/tendermint/tendermint/types"
)

const (
	// Key parts
	_IBC        = "ibc"
	_BLOCKCHAIN = "blockchain"
	_GENESIS    = "genesis"
	_STATE      = "state"
	_HEADER     = "header"
	_EGRESS     = "egress"
	_INGRESS    = "ingress"
	_CONNECTION = "connection"
)

type IBCPluginState struct {
	// @[:ibc, :blockchain, :genesis, ChainID] <~ BlockchainGenesis
	// @[:ibc, :blockchain, :state, ChainID] <~ BlockchainState
	// @[:ibc, :blockchain, :header, ChainID, Height] <~ tm.Header
	// @[:ibc, :egress, Src, Dst, Sequence] <~ Packet
	// @[:ibc, :ingress, Dst, Src, Sequence] <~ Packet
	// @[:ibc, :connection, Src, Dst] <~ Connection # TODO - keep connection state
}

type BlockchainGenesis struct {
	ChainID string
	Genesis string
}

type BlockchainState struct {
	ChainID         string
	Validators      []*tm.Validator
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
	IBCTxTypePacketCreate  = byte(0x03)
	IBCTxTypePacketPost    = byte(0x04)

	IBCCodeEncodingError       = abci.CodeType(1001)
	IBCCodeChainAlreadyExists  = abci.CodeType(1002)
	IBCCodePacketAlreadyExists = abci.CodeType(1003)
	IBCCodeUnknownHeight       = abci.CodeType(1004)
	IBCCodeInvalidCommit       = abci.CodeType(1005)
	IBCCodeInvalidProof        = abci.CodeType(1006)
)

var _ = wire.RegisterInterface(
	struct{ IBCTx }{},
	wire.ConcreteType{IBCRegisterChainTx{}, IBCTxTypeRegisterChain},
	wire.ConcreteType{IBCUpdateChainTx{}, IBCTxTypeUpdateChain},
	wire.ConcreteType{IBCPacketCreateTx{}, IBCTxTypePacketCreate},
	wire.ConcreteType{IBCPacketPostTx{}, IBCTxTypePacketPost},
)

type IBCTx interface {
	AssertIsIBCTx()
	ValidateBasic() abci.Result
}

func (IBCRegisterChainTx) AssertIsIBCTx() {}
func (IBCUpdateChainTx) AssertIsIBCTx()   {}
func (IBCPacketCreateTx) AssertIsIBCTx()  {}
func (IBCPacketPostTx) AssertIsIBCTx()    {}

type IBCRegisterChainTx struct {
	BlockchainGenesis
}

func (IBCRegisterChainTx) ValidateBasic() (res abci.Result) {
	// TODO - validate
	return
}

type IBCUpdateChainTx struct {
	Header tm.Header
	Commit tm.Commit
	// TODO: NextValidators
}

func (IBCUpdateChainTx) ValidateBasic() (res abci.Result) {
	// TODO - validate
	return
}

type IBCPacketCreateTx struct {
	Packet
}

func (IBCPacketCreateTx) ValidateBasic() (res abci.Result) {
	// TODO - validate
	return
}

type IBCPacketPostTx struct {
	FromChainID     string // The immediate source of the packet, not always Packet.SrcChainID
	FromChainHeight uint64 // The block height in which Packet was committed, to check Proof
	Packet
	Proof *merkle.IAVLProof
}

func (IBCPacketPostTx) ValidateBasic() (res abci.Result) {
	// TODO - validate
	return
}

//--------------------------------------------------------------------------------

type IBCPlugin struct {
}

func (ibc *IBCPlugin) Name() string {
	return "IBC"
}

func (ibc *IBCPlugin) StateKey() []byte {
	return []byte("IBCPlugin.State")
}

func New() *IBCPlugin {
	return &IBCPlugin{}
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
	res = tx.ValidateBasic()
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
	case IBCPacketCreateTx:
		sm.runPacketCreateTx(tx)
	case IBCPacketPostTx:
		sm.runPacketPostTx(tx)
	}

	return sm.res
}

type IBCStateMachine struct {
	store types.KVStore
	ctx   types.CallContext
	res   abci.Result
}

func (sm *IBCStateMachine) runRegisterChainTx(tx IBCRegisterChainTx) {
	chainGenKey := toKey(_IBC, _BLOCKCHAIN, _GENESIS, tx.ChainID)
	chainStateKey := toKey(_IBC, _BLOCKCHAIN, _STATE, tx.ChainID)
	chainGen := tx.BlockchainGenesis

	// Parse genesis
	var chainGenDoc = &tm.GenesisDoc{}
	var err error
	wire.ReadJSONPtr(&chainGenDoc, []byte(chainGen.Genesis), &err)
	if err != nil {
		sm.res.Code = IBCCodeEncodingError
		sm.res.Log = "Genesis doc couldn't be parsed: " + err.Error()
		return
	}

	// Make sure chainGen doesn't already exist
	if exists(sm.store, chainGenKey) {
		sm.res.Code = IBCCodeChainAlreadyExists
		sm.res.Log = "Already exists"
		return
	}

	// Save new BlockchainGenesis
	save(sm.store, chainGenKey, chainGen)

	// Create new BlockchainState
	chainState := BlockchainState{
		ChainID:         chainGenDoc.ChainID,
		Validators:      make([]*tm.Validator, len(chainGenDoc.Validators)),
		LastBlockHash:   nil,
		LastBlockHeight: 0,
	}
	// Make validators slice
	for i, val := range chainGenDoc.Validators {
		pubKey := val.PubKey
		address := pubKey.Address()
		chainState.Validators[i] = &tm.Validator{
			Address:     address,
			PubKey:      pubKey,
			VotingPower: val.Amount,
		}
	}

	// Save new BlockchainState
	save(sm.store, chainStateKey, chainState)
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
		sm.res = abci.ErrInternalError.AppendLog(cmn.Fmt("Loading ChainState: %v", err.Error()))
		return
	}
	if !exists {
		sm.res = abci.ErrInternalError.AppendLog(cmn.Fmt("Missing ChainState"))
		return
	}

	// Check commit against last known state & validators
	err = verifyCommit(chainState, &tx.Header, &tx.Commit)
	if err != nil {
		sm.res.Code = IBCCodeInvalidCommit
		sm.res.Log = cmn.Fmt("Invalid Commit: %v", err.Error())
		return
	}

	// Store header
	headerKey := toKey(_IBC, _BLOCKCHAIN, _HEADER, chainID, cmn.Fmt("%v", tx.Header.Height))
	save(sm.store, headerKey, tx.Header)

	// Update chainState
	chainState.LastBlockHash = tx.Header.Hash()
	chainState.LastBlockHeight = uint64(tx.Header.Height)

	// Store chainState
	save(sm.store, chainStateKey, chainState)
}

func (sm *IBCStateMachine) runPacketCreateTx(tx IBCPacketCreateTx) {
	packet := tx.Packet
	packetKey := toKey(_IBC, _EGRESS,
		packet.SrcChainID,
		packet.DstChainID,
		cmn.Fmt("%v", packet.Sequence),
	)
	// Make sure packet doesn't already exist
	if exists(sm.store, packetKey) {
		sm.res.Code = IBCCodePacketAlreadyExists
		// TODO: .AppendLog() does not update sm.res
		sm.res.Log = "Already exists"
		return
	}
	// Save new Packet
	save(sm.store, packetKey, packet)
}

func (sm *IBCStateMachine) runPacketPostTx(tx IBCPacketPostTx) {
	packet := tx.Packet
	packetKeyEgress := toKey(_IBC, _EGRESS,
		packet.SrcChainID,
		packet.DstChainID,
		cmn.Fmt("%v", packet.Sequence),
	)
	packetKeyIngress := toKey(_IBC, _INGRESS,
		packet.DstChainID,
		packet.SrcChainID,
		cmn.Fmt("%v", packet.Sequence),
	)
	headerKey := toKey(_IBC, _BLOCKCHAIN, _HEADER,
		tx.FromChainID,
		cmn.Fmt("%v", tx.FromChainHeight),
	)

	// Make sure packet doesn't already exist
	if exists(sm.store, packetKeyIngress) {
		sm.res.Code = IBCCodePacketAlreadyExists
		sm.res.Log = "Already exists"
		return
	}

	// Save new Packet
	save(sm.store, packetKeyIngress, packet)

	// Load Header and make sure it exists
	// If it exists, we already checked a valid commit for it in UpdateChainTx
	var header tm.Header
	exists, err := load(sm.store, headerKey, &header)
	if err != nil {
		sm.res = abci.ErrInternalError.AppendLog(cmn.Fmt("Loading Header: %v", err.Error()))
		return
	}
	if !exists {
		sm.res.Code = IBCCodeUnknownHeight
		sm.res.Log = cmn.Fmt("Loading Header: Unknown height")
		return
	}

	proof := tx.Proof
	if proof == nil {
		sm.res.Code = IBCCodeInvalidProof
		sm.res.Log = "Proof is nil"
		return
	}
	packetBytes := wire.BinaryBytes(packet)

	// Make sure packet's proof matches given (packet, key, blockhash)
	ok := proof.Verify(packetKeyEgress, packetBytes, header.AppHash)
	if !ok {
		sm.res.Code = IBCCodeInvalidProof
		sm.res.Log = "Proof is invalid"
		return
	}

	return
}

func (ibc *IBCPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
}

func (cp *IBCPlugin) BeginBlock(store types.KVStore, hash []byte, header *abci.Header) {
}

func (cp *IBCPlugin) EndBlock(store types.KVStore, height uint64) (res abci.ResponseEndBlock) {
	return
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
				cmn.Fmt("Error decoding key 0x%X = 0x%X: %v", key, value, err.Error()),
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

// NOTE: Commit's votes include ValidatorAddress, so can be matched up
// against chainState.Validators, even if the validator set had changed.
// For the purpose of the demo, we assume that the validator set hadn't changed,
// though we should check that explicitly.
func verifyCommit(chainState BlockchainState, header *tm.Header, commit *tm.Commit) error {

	// Ensure that chainState and header ChainID match.
	if chainState.ChainID != header.ChainID {
		return errors.New(cmn.Fmt("Expected header.ChainID %v, got %v", chainState.ChainID, header.ChainID))
	}
	// Ensure things aren't empty
	if len(chainState.Validators) == 0 {
		return errors.New(cmn.Fmt("Blockchain has no validators")) // NOTE: Why would this happen?
	}
	if len(commit.Precommits) == 0 {
		return errors.New(cmn.Fmt("Commit has no signatures"))
	}
	chainID := chainState.ChainID
	vals := chainState.Validators
	valSet := tm.NewValidatorSet(vals)
	blockID := commit.Precommits[0].BlockID // XXX: incorrect

	// NOTE: Currently this only works with the exact same validator set.
	// Not this, but perhaps "ValidatorSet.VerifyCommitAny" should expose
	// the functionality to verify commits even after validator changes.
	err := valSet.VerifyCommit(chainID, blockID, header.Height, commit)
	if err != nil {
		return err
	}

	// Ensure the committed blockID matches the header
	if !bytes.Equal(header.Hash(), blockID.Hash) {
		return errors.New(cmn.Fmt("blockID.Hash (%X) does not match header.Hash (%X)", blockID.Hash, header.Hash()))
	}

	// All ok!
	return nil
}

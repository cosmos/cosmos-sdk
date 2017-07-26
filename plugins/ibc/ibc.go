package ibc

// import (
// 	"bytes"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"net/url"
// 	"strconv"
// 	"strings"

// 	abci "github.com/tendermint/abci/types"
// 	"github.com/tendermint/go-wire"
// 	merkle "github.com/tendermint/merkleeyes/iavl"
// 	cmn "github.com/tendermint/tmlibs/common"

// 	"github.com/tendermint/basecoin/types"
// 	tm "github.com/tendermint/tendermint/types"
// )

// const (
// 	// Key parts
// 	_IBC        = "ibc"
// 	_BLOCKCHAIN = "blockchain"
// 	_GENESIS    = "genesis"
// 	_STATE      = "state"
// 	_HEADER     = "header"
// 	_EGRESS     = "egress"
// 	_INGRESS    = "ingress"
// 	_CONNECTION = "connection"
// )

// type IBCPluginState struct {
// 	// @[:ibc, :blockchain, :genesis, ChainID] <~ BlockchainGenesis
// 	// @[:ibc, :blockchain, :state, ChainID] <~ BlockchainState
// 	// @[:ibc, :blockchain, :header, ChainID, Height] <~ tm.Header
// 	// @[:ibc, :egress, Src, Dst, Sequence] <~ Packet
// 	// @[:ibc, :ingress, Dst, Src, Sequence] <~ Packet
// 	// @[:ibc, :connection, Src, Dst] <~ Connection # TODO - keep connection state
// }

// type BlockchainGenesis struct {
// 	ChainID string
// 	Genesis string
// }

// type BlockchainState struct {
// 	ChainID         string
// 	Validators      []*tm.Validator
// 	LastBlockHash   []byte
// 	LastBlockHeight uint64
// }

// type Packet struct {
// 	SrcChainID string
// 	DstChainID string
// 	Sequence   uint64
// 	Type       string // redundant now that Type() is a method on Payload ?
// 	Payload    Payload
// }

// func NewPacket(src, dst string, seq uint64, payload Payload) Packet {
// 	return Packet{
// 		SrcChainID: src,
// 		DstChainID: dst,
// 		Sequence:   seq,
// 		Type:       payload.Type(),
// 		Payload:    payload,
// 	}
// }

// // GetSequenceNumber gets the sequence number for packets being sent from the src chain to the dst chain.
// // The sequence number counts how many packets have been sent.
// // The next packet must include the latest sequence number.
// func GetSequenceNumber(store state.SimpleDB, src, dst string) uint64 {
// 	sequenceKey := toKey(_IBC, _EGRESS, src, dst)
// 	seqBytes := store.Get(sequenceKey)
// 	if seqBytes == nil {
// 		return 0
// 	}
// 	seq, err := strconv.ParseUint(string(seqBytes), 10, 64)
// 	if err != nil {
// 		cmn.PanicSanity(err.Error())
// 	}
// 	return seq
// }

// // SetSequenceNumber sets the sequence number for packets being sent from the src chain to the dst chain
// func SetSequenceNumber(store state.SimpleDB, src, dst string, seq uint64) {
// 	sequenceKey := toKey(_IBC, _EGRESS, src, dst)
// 	store.Set(sequenceKey, []byte(strconv.FormatUint(seq, 10)))
// }

// // SaveNewIBCPacket creates an IBC packet with the given payload from the src chain to the dst chain
// // using the correct sequence number. It also increments the sequence number by 1
// func SaveNewIBCPacket(state state.SimpleDB, src, dst string, payload Payload) {
// 	// fetch sequence number and increment by 1
// 	seq := GetSequenceNumber(state, src, dst)
// 	SetSequenceNumber(state, src, dst, seq+1)

// 	// save ibc packet
// 	packetKey := toKey(_IBC, _EGRESS, src, dst, cmn.Fmt("%v", seq))
// 	packet := NewPacket(src, dst, uint64(seq), payload)
// 	save(state, packetKey, packet)
// }

// func GetIBCPacket(state state.SimpleDB, src, dst string, seq uint64) (Packet, error) {
// 	packetKey := toKey(_IBC, _EGRESS, src, dst, cmn.Fmt("%v", seq))
// 	packetBytes := state.Get(packetKey)

// 	var packet Packet
// 	err := wire.ReadBinaryBytes(packetBytes, &packet)
// 	return packet, err
// }

// //--------------------------------------------------------------------------------

// const (
// 	PayloadTypeBytes = byte(0x01)
// 	PayloadTypeCoins = byte(0x02)
// )

// var _ = wire.RegisterInterface(
// 	struct{ Payload }{},
// 	wire.ConcreteType{DataPayload{}, PayloadTypeBytes},
// 	wire.ConcreteType{CoinsPayload{}, PayloadTypeCoins},
// )

// type Payload interface {
// 	AssertIsPayload()
// 	Type() string
// 	ValidateBasic() abci.Result
// }

// func (DataPayload) AssertIsPayload()  {}
// func (CoinsPayload) AssertIsPayload() {}

// type DataPayload []byte

// func (p DataPayload) Type() string {
// 	return "data"
// }

// func (p DataPayload) ValidateBasic() abci.Result {
// 	return abci.OK
// }

// type CoinsPayload struct {
// 	Address []byte
// 	Coins   coin.Coins
// }

// func (p CoinsPayload) Type() string {
// 	return "coin"
// }

// func (p CoinsPayload) ValidateBasic() abci.Result {
// 	// TODO: validate
// 	return abci.OK
// }

// //--------------------------------------------------------------------------------

// const (
// 	IBCTxTypeRegisterChain = byte(0x01)
// 	IBCTxTypeUpdateChain   = byte(0x02)
// 	IBCTxTypePacketCreate  = byte(0x03)
// 	IBCTxTypePacketPost    = byte(0x04)

// 	IBCCodeEncodingError       = abci.CodeType(1001)
// 	IBCCodeChainAlreadyExists  = abci.CodeType(1002)
// 	IBCCodePacketAlreadyExists = abci.CodeType(1003)
// 	IBCCodeUnknownHeight       = abci.CodeType(1004)
// 	IBCCodeInvalidCommit       = abci.CodeType(1005)
// 	IBCCodeInvalidProof        = abci.CodeType(1006)
// )

// var _ = wire.RegisterInterface(
// 	struct{ IBCTx }{},
// 	wire.ConcreteType{IBCRegisterChainTx{}, IBCTxTypeRegisterChain},
// 	wire.ConcreteType{IBCUpdateChainTx{}, IBCTxTypeUpdateChain},
// 	wire.ConcreteType{IBCPacketCreateTx{}, IBCTxTypePacketCreate},
// 	wire.ConcreteType{IBCPacketPostTx{}, IBCTxTypePacketPost},
// )

// type IBCTx interface {
// 	AssertIsIBCTx()
// 	ValidateBasic() abci.Result
// }

// func (IBCRegisterChainTx) AssertIsIBCTx() {}
// func (IBCUpdateChainTx) AssertIsIBCTx()   {}
// func (IBCPacketCreateTx) AssertIsIBCTx()  {}
// func (IBCPacketPostTx) AssertIsIBCTx()    {}

// type IBCRegisterChainTx struct {
// 	BlockchainGenesis
// }

// func (IBCRegisterChainTx) ValidateBasic() (res abci.Result) {
// 	// TODO - validate
// 	return
// }

// type IBCUpdateChainTx struct {
// 	Header tm.Header
// 	Commit tm.Commit
// 	// TODO: NextValidators
// }

// func (IBCUpdateChainTx) ValidateBasic() (res abci.Result) {
// 	// TODO - validate
// 	return
// }

// type IBCPacketCreateTx struct {
// 	Packet
// }

// func (IBCPacketCreateTx) ValidateBasic() (res abci.Result) {
// 	// TODO - validate
// 	return
// }

// type IBCPacketPostTx struct {
// 	FromChainID     string // The immediate source of the packet, not always Packet.SrcChainID
// 	FromChainHeight uint64 // The block height in which Packet was committed, to check Proof
// 	Packet
// 	Proof *merkle.IAVLProof
// }

// func (IBCPacketPostTx) ValidateBasic() (res abci.Result) {
// 	// TODO - validate
// 	return
// }

// //--------------------------------------------------------------------------------

// type IBCPlugin struct {
// }

// func (ibc *IBCPlugin) Name() string {
// 	return "IBC"
// }

// func (ibc *IBCPlugin) StateKey() []byte {
// 	return []byte("IBCPlugin.State")
// }

// func New() *IBCPlugin {
// 	return &IBCPlugin{}
// }

// func (ibc *IBCPlugin) SetOption(store state.SimpleDB, key string, value string) (log string) {
// 	return ""
// }

// func (ibc *IBCPlugin) RunTx(store state.SimpleDB, ctx types.CallContext, txBytes []byte) (res abci.Result) {
// 	// Decode tx
// 	var tx IBCTx
// 	err := wire.ReadBinaryBytes(txBytes, &tx)
// 	if err != nil {
// 		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
// 	}

// 	// Validate tx
// 	res = tx.ValidateBasic()
// 	if res.IsErr() {
// 		return res.PrependLog("ValidateBasic Failed: ")
// 	}

// 	// TODO - Check whether sufficient funds

// 	defer func() {
// 		// TODO - Refund any remaining funds left over
// 		// e.g. !ctx.Coins.Minus(tx.Fee).IsZero()
// 		// ctx.CallerAccount is synced w/ store, so just modify that and store it.
// 		// NOTE: We should use the CallContext to store fund/refund information.
// 	}()

// 	sm := &IBCStateMachine{store, ctx, abci.OK}

// 	switch tx := tx.(type) {
// 	case IBCRegisterChainTx:
// 		sm.runRegisterChainTx(tx)
// 	case IBCUpdateChainTx:
// 		sm.runUpdateChainTx(tx)
// 	case IBCPacketCreateTx:
// 		sm.runPacketCreateTx(tx)
// 	case IBCPacketPostTx:
// 		sm.runPacketPostTx(tx)
// 	}

// 	return sm.res
// }

// type IBCStateMachine struct {
// 	store state.SimpleDB
// 	ctx   types.CallContext
// 	res   abci.Result
// }

// func (sm *IBCStateMachine) runRegisterChainTx(tx IBCRegisterChainTx) {
// 	chainGenKey := toKey(_IBC, _BLOCKCHAIN, _GENESIS, tx.ChainID)
// 	chainStateKey := toKey(_IBC, _BLOCKCHAIN, _STATE, tx.ChainID)
// 	chainGen := tx.BlockchainGenesis

// 	// Parse genesis
// 	chainGenDoc := new(tm.GenesisDoc)
// 	err := json.Unmarshal([]byte(chainGen.Genesis), chainGenDoc)
// 	if err != nil {
// 		sm.res.Code = IBCCodeEncodingError
// 		sm.res.Log = "Genesis doc couldn't be parsed: " + err.Error()
// 		return
// 	}

// 	// Make sure chainGen doesn't already exist
// 	if exists(sm.store, chainGenKey) {
// 		sm.res.Code = IBCCodeChainAlreadyExists
// 		sm.res.Log = "Already exists"
// 		return
// 	}

// 	// Save new BlockchainGenesis
// 	save(sm.store, chainGenKey, chainGen)

// 	// Create new BlockchainState
// 	chainState := BlockchainState{
// 		ChainID:         chainGenDoc.ChainID,
// 		Validators:      make([]*tm.Validator, len(chainGenDoc.Validators)),
// 		LastBlockHash:   nil,
// 		LastBlockHeight: 0,
// 	}
// 	// Make validators slice
// 	for i, val := range chainGenDoc.Validators {
// 		pubKey := val.PubKey
// 		address := pubKey.Address()
// 		chainState.Validators[i] = &tm.Validator{
// 			Address:     address,
// 			PubKey:      pubKey,
// 			VotingPower: val.Amount,
// 		}
// 	}

// 	// Save new BlockchainState
// 	save(sm.store, chainStateKey, chainState)
// }

// func (sm *IBCStateMachine) runUpdateChainTx(tx IBCUpdateChainTx) {
// 	chainID := tx.Header.ChainID
// 	chainStateKey := toKey(_IBC, _BLOCKCHAIN, _STATE, chainID)

// 	// Make sure chainState exists
// 	if !exists(sm.store, chainStateKey) {
// 		return // Chain does not exist, do nothing
// 	}

// 	// Load latest chainState
// 	var chainState BlockchainState
// 	exists, err := load(sm.store, chainStateKey, &chainState)
// 	if err != nil {
// 		sm.res = abci.ErrInternalError.AppendLog(cmn.Fmt("Loading ChainState: %v", err.Error()))
// 		return
// 	}
// 	if !exists {
// 		sm.res = abci.ErrInternalError.AppendLog(cmn.Fmt("Missing ChainState"))
// 		return
// 	}

// 	// Check commit against last known state & validators
// 	err = verifyCommit(chainState, &tx.Header, &tx.Commit)
// 	if err != nil {
// 		sm.res.Code = IBCCodeInvalidCommit
// 		sm.res.Log = cmn.Fmt("Invalid Commit: %v", err.Error())
// 		return
// 	}

// 	// Store header
// 	headerKey := toKey(_IBC, _BLOCKCHAIN, _HEADER, chainID, cmn.Fmt("%v", tx.Header.Height))
// 	save(sm.store, headerKey, tx.Header)

// 	// Update chainState
// 	chainState.LastBlockHash = tx.Header.Hash()
// 	chainState.LastBlockHeight = uint64(tx.Header.Height)

// 	// Store chainState
// 	save(sm.store, chainStateKey, chainState)
// }

// func (sm *IBCStateMachine) runPacketCreateTx(tx IBCPacketCreateTx) {
// 	packet := tx.Packet
// 	packetKey := toKey(_IBC, _EGRESS,
// 		packet.SrcChainID,
// 		packet.DstChainID,
// 		cmn.Fmt("%v", packet.Sequence),
// 	)
// 	// Make sure packet doesn't already exist
// 	if exists(sm.store, packetKey) {
// 		sm.res.Code = IBCCodePacketAlreadyExists
// 		// TODO: .AppendLog() does not update sm.res
// 		sm.res.Log = "Already exists"
// 		return
// 	}

// 	// Execute the payload
// 	switch payload := tx.Packet.Payload.(type) {
// 	case DataPayload:
// 		// do nothing
// 	case CoinsPayload:
// 		// ensure enough coins were sent in tx to cover the payload coins
// 		if !sm.ctx.Coins.IsGTE(payload.Coins) {
// 			sm.res.Code = abci.CodeType_InsufficientFunds
// 			sm.res.Log = fmt.Sprintf("Not enough funds sent in tx (%v) to send %v via IBC", sm.ctx.Coins, payload.Coins)
// 			return
// 		}

// 		// deduct coins from context
// 		sm.ctx.Coins = sm.ctx.Coins.Minus(payload.Coins)
// 	}

// 	// Save new Packet
// 	save(sm.store, packetKey, packet)

// 	// set the sequence number
// 	SetSequenceNumber(sm.store, packet.SrcChainID, packet.DstChainID, packet.Sequence)
// }

// func (sm *IBCStateMachine) runPacketPostTx(tx IBCPacketPostTx) {
// 	packet := tx.Packet
// 	packetKeyEgress := toKey(_IBC, _EGRESS,
// 		packet.SrcChainID,
// 		packet.DstChainID,
// 		cmn.Fmt("%v", packet.Sequence),
// 	)
// 	packetKeyIngress := toKey(_IBC, _INGRESS,
// 		packet.DstChainID,
// 		packet.SrcChainID,
// 		cmn.Fmt("%v", packet.Sequence),
// 	)
// 	headerKey := toKey(_IBC, _BLOCKCHAIN, _HEADER,
// 		tx.FromChainID,
// 		cmn.Fmt("%v", tx.FromChainHeight),
// 	)

// 	// Make sure packet doesn't already exist
// 	if exists(sm.store, packetKeyIngress) {
// 		sm.res.Code = IBCCodePacketAlreadyExists
// 		sm.res.Log = "Already exists"
// 		return
// 	}

// 	// Save new Packet (just for fun)
// 	save(sm.store, packetKeyIngress, packet)

// 	// Load Header and make sure it exists
// 	// If it exists, we already checked a valid commit for it in UpdateChainTx
// 	var header tm.Header
// 	exists, err := load(sm.store, headerKey, &header)
// 	if err != nil {
// 		sm.res = abci.ErrInternalError.AppendLog(cmn.Fmt("Loading Header: %v", err.Error()))
// 		return
// 	}
// 	if !exists {
// 		sm.res.Code = IBCCodeUnknownHeight
// 		sm.res.Log = cmn.Fmt("Loading Header: Unknown height")
// 		return
// 	}

// 	proof := tx.Proof
// 	if proof == nil {
// 		sm.res.Code = IBCCodeInvalidProof
// 		sm.res.Log = "Proof is nil"
// 		return
// 	}
// 	packetBytes := wire.BinaryBytes(packet)

// 	// Make sure packet's proof matches given (packet, key, blockhash)
// 	ok := proof.Verify(packetKeyEgress, packetBytes, header.AppHash)
// 	if !ok {
// 		sm.res.Code = IBCCodeInvalidProof
// 		sm.res.Log = fmt.Sprintf("Proof is invalid. key: %s; packetByes %X; header %v; proof %v", packetKeyEgress, packetBytes, header, proof)
// 		return
// 	}

// 	// Execute payload
// 	switch payload := packet.Payload.(type) {
// 	case DataPayload:
// 		// do nothing
// 	case CoinsPayload:
// 		// Add coins to destination account
// 		acc := types.GetAccount(sm.store, payload.Address)
// 		if acc == nil {
// 			acc = &types.Account{}
// 		}
// 		acc.Balance = acc.Balance.Plus(payload.Coins)
// 		types.SetAccount(sm.store, payload.Address, acc)
// 	}

// 	return
// }

// func (ibc *IBCPlugin) InitChain(store state.SimpleDB, vals []*abci.Validator) {
// }

// func (cp *IBCPlugin) BeginBlock(store state.SimpleDB, hash []byte, header *abci.Header) {
// }

// func (cp *IBCPlugin) EndBlock(store state.SimpleDB, height uint64) (res abci.ResponseEndBlock) {
// 	return
// }

// //--------------------------------------------------------------------------------
// // TODO: move to utils

// // Returns true if exists, false if nil.
// func exists(store state.SimpleDB, key []byte) (exists bool) {
// 	value := store.Get(key)
// 	return len(value) > 0
// }

// // Load bytes from store by reading value for key and read into ptr.
// // Returns true if exists, false if nil.
// // Returns err if decoding error.
// func load(store state.SimpleDB, key []byte, ptr interface{}) (exists bool, err error) {
// 	value := store.Get(key)
// 	if len(value) > 0 {
// 		err = wire.ReadBinaryBytes(value, ptr)
// 		if err != nil {
// 			return true, errors.New(
// 				cmn.Fmt("Error decoding key 0x%X = 0x%X: %v", key, value, err.Error()),
// 			)
// 		}
// 		return true, nil
// 	} else {
// 		return false, nil
// 	}
// }

// // Save bytes to store by writing obj's go-wire binary bytes.
// func save(store state.SimpleDB, key []byte, obj interface{}) {
// 	store.Set(key, wire.BinaryBytes(obj))
// }

// // Key parts are URL escaped and joined with ','
// func toKey(parts ...string) []byte {
// 	escParts := make([]string, len(parts))
// 	for i, part := range parts {
// 		escParts[i] = url.QueryEscape(part)
// 	}
// 	return []byte(strings.Join(escParts, ","))
// }

// // NOTE: Commit's votes include ValidatorAddress, so can be matched up
// // against chainState.Validators, even if the validator set had changed.
// // For the purpose of the demo, we assume that the validator set hadn't changed,
// // though we should check that explicitly.
// func verifyCommit(chainState BlockchainState, header *tm.Header, commit *tm.Commit) error {

// 	// Ensure that chainState and header ChainID match.
// 	if chainState.ChainID != header.ChainID {
// 		return errors.New(cmn.Fmt("Expected header.ChainID %v, got %v", chainState.ChainID, header.ChainID))
// 	}
// 	// Ensure things aren't empty
// 	if len(chainState.Validators) == 0 {
// 		return errors.New(cmn.Fmt("Blockchain has no validators")) // NOTE: Why would this happen?
// 	}
// 	if len(commit.Precommits) == 0 {
// 		return errors.New(cmn.Fmt("Commit has no signatures"))
// 	}
// 	chainID := chainState.ChainID
// 	vals := chainState.Validators
// 	valSet := tm.NewValidatorSet(vals)

// 	var blockID tm.BlockID
// 	for _, pc := range commit.Precommits {
// 		// XXX: incorrect. we want the one for +2/3, not just the first one
// 		if pc != nil {
// 			blockID = pc.BlockID
// 		}
// 	}
// 	if blockID.IsZero() {
// 		return errors.New("All precommits are nil!")
// 	}

// 	// NOTE: Currently this only works with the exact same validator set.
// 	// Not this, but perhaps "ValidatorSet.VerifyCommitAny" should expose
// 	// the functionality to verify commits even after validator changes.
// 	err := valSet.VerifyCommit(chainID, blockID, header.Height, commit)
// 	if err != nil {
// 		return err
// 	}

// 	// Ensure the committed blockID matches the header
// 	if !bytes.Equal(header.Hash(), blockID.Hash) {
// 		return errors.New(cmn.Fmt("blockID.Hash (%X) does not match header.Hash (%X)", blockID.Hash, header.Hash()))
// 	}

// 	// All ok!
// 	return nil
// }

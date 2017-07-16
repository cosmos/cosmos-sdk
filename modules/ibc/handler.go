package ibc

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

const (
	// NameIBC is the name of this module
	NameIBC = "ibc"
)

var (
	allowIBC = []byte{0x42, 0xbe, 0xef, 0x1}
)

// AllowIBC is the special code that an app must set to
// enable sending IBC packets for this app-type
func AllowIBC(app string) basecoin.Actor {
	return basecoin.Actor{App: app, Address: allowIBC}
}

// Handler allows us to update the chain state or create a packet
//
// TODO: require auth for registration, the authorized actor (or role)
// should be defined in the handler, and set via SetOption
type Handler struct {
	// TODO: add option to set who can permit registration and store it
	basecoin.NopOption
}

var _ basecoin.Handler = Handler{}

// NewHandler makes a role handler to create roles
func NewHandler() Handler {
	return Handler{}
}

// Name - return name space
func (Handler) Name() string {
	return NameIBC
}

// CheckTx verifies the packet is formated correctly, and has the proper sequence
// for a registered chain
func (h Handler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case RegisterChainTx:
		return h.initSeed(ctx, store, t)
	case UpdateChainTx:
		return h.updateSeed(ctx, store, t)
	case CreatePacketTx:
		return h.createPacket(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// DeliverTx verifies all signatures on the tx and updated the chain state
// apropriately
func (h Handler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case RegisterChainTx:
		// TODO: do we want some permissioning for this???
		return h.initSeed(ctx, store, t)
	case UpdateChainTx:
		return h.updateSeed(ctx, store, t)
	case CreatePacketTx:
		return h.createPacket(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// initSeed imports the first seed for this chain and accepts it as the root of trust
func (h Handler) initSeed(ctx basecoin.Context, store state.KVStore,
	t RegisterChainTx) (res basecoin.Result, err error) {

	chainID := t.ChainID()
	s := NewChainSet(store)
	err = s.Register(chainID, ctx.BlockHeight(), t.Seed.Height())
	if err != nil {
		return res, err
	}

	space := stack.PrefixedStore(chainID, store)
	provider := newDBProvider(space)
	err = provider.StoreSeed(t.Seed)
	return res, err
}

// updateSeed checks the seed against the existing chain data and rejects it if it
// doesn't fit (or no chain data)
func (h Handler) updateSeed(ctx basecoin.Context, store state.KVStore,
	t UpdateChainTx) (res basecoin.Result, err error) {

	chainID := t.ChainID()
	if !NewChainSet(store).Exists([]byte(chainID)) {
		return res, ErrNotRegistered(chainID)
	}

	// load the certifier for this chain
	seed := t.Seed
	space := stack.PrefixedStore(chainID, store)
	cert, err := newCertifier(space, chainID, seed.Height())
	if err != nil {
		return res, err
	}

	// this will import the seed if it is valid in the current context
	err = cert.Update(seed.Checkpoint, seed.Validators)
	return res, err
}

// createPacket makes sure all permissions are good and the destination
// chain is registed.  If so, it appends it to the outgoing queue
func (h Handler) createPacket(ctx basecoin.Context, store state.KVStore,
	t CreatePacketTx) (res basecoin.Result, err error) {

	// make sure the chain is registed
	dest := t.DestChain
	if !NewChainSet(store).Exists([]byte(dest)) {
		return res, ErrNotRegistered(dest)
	}

	// make sure we have the special IBC permission
	mod, err := t.Tx.GetKind()
	if err != nil {
		return res, err
	}
	if !ctx.HasPermission(AllowIBC(mod)) {
		return res, ErrNeedsIBCPermission()
	}

	// start making the packet to send
	packet := Packet{
		DestChain:   t.DestChain,
		Tx:          t.Tx,
		Permissions: make([]basecoin.Actor, len(t.Permissions)),
	}

	// make sure we have all the permissions we want to send
	for i, p := range t.Permissions {
		if !ctx.HasPermission(p) {
			return res, ErrCannotSetPermission()
		}
		// add the permission with the current ChainID
		packet.Permissions[i] = p
		packet.Permissions[i].ChainID = ctx.ChainID()
	}

	// now add it to the output queue....
	// TODO: where to store, also set the sequence....
	return res, nil
}

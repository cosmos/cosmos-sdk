package ibc

import (
	"fmt"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

const (
	// NameIBC is the name of this module
	NameIBC = "ibc"
	// OptionRegistrar is the option name to set the actor
	// to handle ibc chain registration
	OptionRegistrar = "registrar"
)

var (
	// Semi-random bytes that shouldn't conflict with keys (20 bytes)
	// or any strings (non-ascii).
	// TODO: consider how to make this more collision-proof....
	allowIBC = []byte{0x42, 0xbe, 0xef, 0x1}
)

// AllowIBC returns a specially crafted Actor that
// enables sending IBC packets for this app type
func AllowIBC(app string) basecoin.Actor {
	return basecoin.Actor{App: app, Address: allowIBC}
}

// Handler updates the chain state or creates an ibc packet
type Handler struct{}

var _ basecoin.Handler = Handler{}

// NewHandler returns a Handler that allows all chains to connect via IBC.
// Set a Registrar via SetOption to restrict it.
func NewHandler() Handler {
	return Handler{}
}

// Name returns name space
func (Handler) Name() string {
	return NameIBC
}

// SetOption sets the registrar for IBC
func (h Handler) SetOption(l log.Logger, store state.KVStore, module, key, value string) (log string, err error) {
	if module != NameIBC {
		return "", errors.ErrUnknownModule(module)
	}
	if key == OptionRegistrar {
		var act basecoin.Actor
		err = data.FromJSON([]byte(value), &act)
		if err != nil {
			return "", err
		}
		// Save the data
		info := HandlerInfo{act}
		info.Save(store)
		return "Success", nil
	}
	return "", errors.ErrUnknownKey(key)
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
		// check permission to attach, do it here, so no permission check
		// by SetOption
		info := LoadInfo(store)
		if !info.Registrar.Empty() && !ctx.HasPermission(info.Registrar) {
			return res, errors.ErrUnauthorized()
		}
		return h.initSeed(ctx, store, t)
	case UpdateChainTx:
		return h.updateSeed(ctx, store, t)
	case CreatePacketTx:
		return h.createPacket(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// DeliverTx verifies all signatures on the tx and updates the chain state
// apropriately
func (h Handler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case RegisterChainTx:
		// check permission to attach, do it here, so no permission check
		// by SetOption
		info := LoadInfo(store)
		if !info.Registrar.Empty() && !ctx.HasPermission(info.Registrar) {
			return res, errors.ErrUnauthorized()
		}
		return h.initSeed(ctx, store, t)
	case UpdateChainTx:
		return h.updateSeed(ctx, store, t)
	case CreatePacketTx:
		return h.createPacket(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// initSeed imports the first seed for this chain and
// accepts it as the root of trust.
//
// only the registrar, if set, is allowed to do this
func (h Handler) initSeed(ctx basecoin.Context, store state.KVStore,
	t RegisterChainTx) (res basecoin.Result, err error) {

	// verify that the header looks reasonable
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
	s := NewChainSet(store)
	if !s.Exists([]byte(chainID)) {
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
	if err != nil {
		return res, ErrInvalidCommit(err)
	}

	// update the tracked height in chain info
	err = s.Update(chainID, t.Seed.Height())
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
	mod, err := t.Tx.GetMod()
	if err != nil {
		return res, err
	}
	if !ctx.HasPermission(AllowIBC(mod)) {
		return res, ErrNeedsIBCPermission()
	}

	// start making the packet to send
	packet := Packet{
		DestChain:   dest,
		Tx:          t.Tx,
		Permissions: make([]basecoin.Actor, len(t.Permissions)),
	}

	// make sure we have all the permissions we want to send
	for i, p := range t.Permissions {
		if !ctx.HasPermission(p) {
			return res, ErrCannotSetPermission()
		}
		// add the permission with the current ChainID
		packet.Permissions[i] = p.WithChain(ctx.ChainID())
	}

	// now add it to the output queue....
	q := OutputQueue(store, dest)
	packet.Sequence = q.Tail()
	q.Push(packet.Bytes())

	res = basecoin.Result{Log: fmt.Sprintf("Packet %s %d", dest, packet.Sequence)}
	return
}

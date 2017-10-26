package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

// Middleware allows us to verify the IBC proof on a packet and
// and if valid, attach this permission to the wrapped packet
type Middleware struct {
	stack.PassInitState
	stack.PassInitValidate
}

var _ stack.Middleware = Middleware{}

// NewMiddleware creates a role-checking middleware
func NewMiddleware() Middleware {
	return Middleware{}
}

// Name - return name space
func (Middleware) Name() string {
	return NameIBC
}

// CheckTx verifies the named chain and height is present, and verifies
// the merkle proof in the packet
func (m Middleware) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	// if it is not a PostPacket, just let it go through
	post, ok := tx.Unwrap().(PostPacketTx)
	if !ok {
		return next.CheckTx(ctx, store, tx)
	}

	// parse this packet and get the ibc-enhanced tx and context
	ictx, itx, err := m.verifyPost(ctx, store, post)
	if err != nil {
		return res, err
	}
	return next.CheckTx(ictx, store, itx)
}

// DeliverTx verifies the named chain and height is present, and verifies
// the merkle proof in the packet
func (m Middleware) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	// if it is not a PostPacket, just let it go through
	post, ok := tx.Unwrap().(PostPacketTx)
	if !ok {
		return next.DeliverTx(ctx, store, tx)
	}

	// parse this packet and get the ibc-enhanced tx and context
	ictx, itx, err := m.verifyPost(ctx, store, post)
	if err != nil {
		return res, err
	}
	return next.DeliverTx(ictx, store, itx)
}

// verifyPost accepts a message bound for this chain...
// TODO: think about relay
func (m Middleware) verifyPost(ctx sdk.Context, store state.SimpleDB,
	tx PostPacketTx) (ictx sdk.Context, itx sdk.Tx, err error) {

	// make sure the chain is registered
	from := tx.FromChainID
	if !NewChainSet(store).Exists([]byte(from)) {
		return ictx, itx, ErrNotRegistered(from)
	}

	// TODO: how to deal with routing/relaying???
	packet := tx.Packet
	if packet.DestChain != ctx.ChainID() {
		return ictx, itx, ErrWrongDestChain(packet.DestChain)
	}

	// verify packet.Permissions all come from the other chain
	if !packet.Permissions.AllHaveChain(tx.FromChainID) {
		return ictx, itx, ErrCannotSetPermission()
	}

	// make sure this sequence number is the next in the list
	q := InputQueue(store, from)
	tail := q.Tail()
	if packet.Sequence < tail {
		return ictx, itx, ErrPacketAlreadyExists()
	}
	if packet.Sequence > tail {
		return ictx, itx, ErrPacketOutOfOrder(tail)
	}

	// look up the referenced header
	space := stack.PrefixedStore(from, store)
	provider := newDBProvider(space)

	// if the query was on height H, the proof is in header H+1
	proofHeight := int(tx.FromChainHeight + 1)
	seed, err := provider.GetExactHeight(proofHeight)
	if err != nil {
		return ictx, itx, err
	}

	// verify the merkle hash....
	root := seed.Header.AppHash
	pBytes := packet.Bytes()
	err = tx.Proof.Verify(tx.Key, pBytes, root)
	if err != nil {
		return ictx, itx, ErrInvalidProofWithReason(err)
	}

	// add to input queue
	q.Push(pBytes)

	// return the wrapped tx along with the extra permissions
	ictx = ctx.WithPermissions(packet.Permissions...)
	itx = packet.Tx
	return
}

package ibc

import (
	"errors"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// Middleware allows us to verify the IBC proof on a packet and
// and if valid, attach this permission to the wrapped packet
type Middleware struct {
	stack.PassOption
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
func (m Middleware) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
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
func (m Middleware) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
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
func (m Middleware) verifyPost(ctx basecoin.Context, store state.KVStore,
	tx PostPacketTx) (ictx basecoin.Context, itx basecoin.Tx, err error) {

	// make sure the chain is registered
	from := tx.FromChainID
	if !NewChainSet(store).Exists([]byte(from)) {
		err = ErrNotRegistered(from)
		return
	}

	// make sure this sequence number is the next in the list
	q := InputQueue(store, from)
	packet := tx.Packet
	if q.Tail() != packet.Sequence {
		err = errors.New("Incorrect sequence number - out of order") // TODO
		return
	}

	// look up the referenced header
	space := stack.PrefixedStore(from, store)
	provider := newDBProvider(space)
	// TODO: GetExactHeight helper?
	seed, err := provider.GetByHeight(int(tx.FromChainHeight))
	if err != nil {
		return ictx, itx, err
	}
	if seed.Height() != int(tx.FromChainHeight) {
		err = errors.New("no such height") // TODO
		return
	}

	// verify the merkle hash....
	root := seed.Header.AppHash
	key := []byte("?????") // TODO!
	tx.Proof.Verify(key, packet.Bytes(), root)

	// TODO: verify packet.Permissions

	// add to input queue
	q.Push(packet.Bytes())

	// return the wrapped tx along with the extra permissions
	ictx = ictx.WithPermissions(packet.Permissions...)
	itx = packet.Tx
	return
}

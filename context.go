package basecoin

import (
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"
)

// Actor abstracts any address that can authorize actions, hold funds,
// or initiate any sort of transaction.
//
// It doesn't just have to be a pubkey on this chain, it could stem from
// another app (like multi-sig account), or even another chain (via IBC)
type Actor struct {
	ChainID string     `json:"chain"` // this is empty unless it comes from a different chain
	App     string     `json:"app"`   // the app that the actor belongs to
	Address data.Bytes `json:"addr"`  // arbitrary app-specific unique id
}

func NewActor(app string, addr []byte) Actor {
	return Actor{App: app, Address: addr}
}

func (a Actor) Bytes() []byte {
	return wire.BinaryBytes(a)
}

// Context is an interface, so we can implement "secure" variants that
// rely on private fields to control the actions
type Context interface {
	// context.Context
	log.Logger
	WithPermissions(perms ...Actor) Context
	HasPermission(perm Actor) bool
	GetPermissions(chain, app string) []Actor
	IsParent(ctx Context) bool
	Reset() Context
	ChainID() string
}

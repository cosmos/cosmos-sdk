package basecoin

import "github.com/tendermint/go-wire/data"

// Actor abstracts any address that can authorize actions, hold funds,
// or initiate any sort of transaction.
//
// It doesn't just have to be a pubkey on this chain, it could stem from
// another app (like multi-sig account), or even another chain (via IBC)
type Actor struct {
	ChainID string     // this is empty unless it comes from a different chain
	App     string     // the app that the actor belongs to
	Address data.Bytes // arbitrary app-specific unique id
}

func NewActor(app string, addr []byte) Actor {
	return Actor{App: app, Address: addr}
}

// Context is an interface, so we can implement "secure" variants that
// rely on private fields to control the actions
type Context interface {
	// context.Context
	WithPermissions(perms ...Actor) Context
	HasPermission(perm Actor) bool
	IsParent(ctx Context) bool
	Reset() Context
}

package basecoin

import (
	"bytes"
	"fmt"
	"sort"

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

// NewActor - create a new actor
func NewActor(app string, addr []byte) Actor {
	return Actor{App: app, Address: addr}
}

// Bytes makes a binary coding, useful for turning this into a key in the store
func (a Actor) Bytes() []byte {
	return wire.BinaryBytes(a)
}

// Equals checks if two actors are the same
func (a Actor) Equals(b Actor) bool {
	return a.ChainID == b.ChainID &&
		a.App == b.App &&
		bytes.Equal(a.Address, b.Address)
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
	BlockHeight() uint64
}

//////////////////////////////// Sort Interface
// USAGE sort.Sort(ByAddress(<actor instance>))

func (a Actor) String() string {
	return fmt.Sprintf("%x", a.Address)
}

// ByAddress implements sort.Interface for []Actor based on
// the Address field.
type ByAddress []Actor

// Verify the sort interface at compile time
var _ sort.Interface = ByAddress{}

func (a ByAddress) Len() int           { return len(a) }
func (a ByAddress) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAddress) Less(i, j int) bool { return bytes.Compare(a[i].Address, a[j].Address) == -1 }

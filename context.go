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

// Empty checks if the actor is not initialized
func (a Actor) Empty() bool {
	return a.ChainID == "" && a.App == "" && len(a.Address) == 0
}

// WithChain creates a copy of the actor with a different chainID
func (a Actor) WithChain(chainID string) (b Actor) {
	b = a
	b.ChainID = chainID
	return
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
// USAGE sort.Sort(ByAll(<actor instance>))

func (a Actor) String() string {
	return fmt.Sprintf("%x", a.Address)
}

// ByAll implements sort.Interface for []Actor.
// It sorts be the ChainID, followed by the App, followed by the Address
type ByAll []Actor

// Verify the sort interface at compile time
var _ sort.Interface = ByAll{}

func (a ByAll) Len() int      { return len(a) }
func (a ByAll) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByAll) Less(i, j int) bool {

	if a[i].ChainID < a[j].ChainID {
		return true
	}
	if a[i].ChainID > a[j].ChainID {
		return false
	}
	if a[i].App < a[j].App {
		return true
	}
	if a[i].App > a[j].App {
		return false
	}
	return bytes.Compare(a[i].Address, a[j].Address) == -1
}

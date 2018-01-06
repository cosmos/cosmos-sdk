package ibc

import (
	"fmt"

	"github.com/tendermint/iavl"
	"github.com/tendermint/light-client/certifiers"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

// MockChain is used to simulate a chain for ibc tests.
// It is able to produce ibc packets and all verification for
// them, but cannot respond to any responses.
type MockChain struct {
	keys    certifiers.ValKeys
	chainID string
	tree    *iavl.Tree
}

// NewMockChain initializes a teststore and test validators
func NewMockChain(chainID string, numKeys int) MockChain {
	return MockChain{
		keys:    certifiers.GenValKeys(numKeys),
		chainID: chainID,
		tree:    iavl.NewTree(0, nil),
	}
}

// GetRegistrationTx returns a valid tx to register this chain
func (m MockChain) GetRegistrationTx(h int) RegisterChainTx {
	fc := genEmptyCommit(m.keys, m.chainID, h, m.tree.Hash(), len(m.keys))
	return RegisterChainTx{fc}
}

// MakePostPacket commits the packet locally and returns the proof,
// in the form of two packets to update the header and prove this packet.
func (m MockChain) MakePostPacket(packet Packet, h int) (
	PostPacketTx, UpdateChainTx) {

	post := makePostPacket(m.tree, packet, m.chainID, h)
	fc := genEmptyCommit(m.keys, m.chainID, h+1, m.tree.Hash(), len(m.keys))
	update := UpdateChainTx{fc}

	return post, update
}

func genEmptyCommit(keys certifiers.ValKeys, chain string, h int,
	appHash []byte, count int) certifiers.FullCommit {

	vals := keys.ToValidators(10, 0)
	return keys.GenFullCommit(chain, h, nil, vals, appHash, 0, count)
}

func makePostPacket(tree *iavl.Tree, packet Packet, fromID string, fromHeight int) PostPacketTx {
	key := []byte(fmt.Sprintf("some-long-prefix-%06d", packet.Sequence))
	tree.Set(key, packet.Bytes())
	_, proof, err := tree.GetWithProof(key)
	if err != nil {
		panic(err)
	}
	if proof == nil {
		panic("wtf?")
	}

	return PostPacketTx{
		FromChainID:     fromID,
		FromChainHeight: uint64(fromHeight),
		Proof:           proof.(*iavl.KeyExistsProof),
		Key:             key,
		Packet:          packet,
	}
}

// AppChain is ready to handle tx
type AppChain struct {
	chainID string
	app     sdk.Handler
	store   state.SimpleDB
	height  int
}

// NewAppChain returns a chain that is ready to respond to tx
func NewAppChain(app sdk.Handler, chainID string) *AppChain {
	return &AppChain{
		chainID: chainID,
		app:     app,
		store:   state.NewMemKVStore(),
		height:  123,
	}
}

// IncrementHeight allows us to jump heights, more than the auto-step
// of 1.  It returns the new height we are at.
func (a *AppChain) IncrementHeight(delta int) int {
	a.height += delta
	return a.height
}

// DeliverTx runs the tx and commits the new tree, incrementing height
// by one.
func (a *AppChain) DeliverTx(tx sdk.Tx, perms ...sdk.Actor) (sdk.DeliverResult, error) {
	ctx := stack.MockContext(a.chainID, uint64(a.height)).WithPermissions(perms...)
	store := a.store.Checkpoint()
	res, err := a.app.DeliverTx(ctx, store, tx)
	if err == nil {
		// commit data on success
		a.store.Commit(store)
	}
	return res, err
}

// Update is a shortcut to DeliverTx with this.  Also one return value
// to test inline
func (a *AppChain) Update(tx UpdateChainTx) error {
	_, err := a.DeliverTx(tx.Wrap())
	return err
}

// InitState sets the option on our app
func (a *AppChain) InitState(mod, key, value string) (string, error) {
	return a.app.InitState(log.NewNopLogger(), a.store, mod, key, value)
}

// GetStore is used to get the app-specific sub-store
func (a *AppChain) GetStore(app string) state.SimpleDB {
	return stack.PrefixedStore(app, a.store)
}

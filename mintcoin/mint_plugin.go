package mintcoin

import (
	"encoding/hex"
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
)

const (
	AddBanker    = "add"
	RemoveBanker = "remove"
)

// MintPlugin is a plugin, storing all state prefixed with it's unique name
type MintPlugin struct {
	name string
}

func New(name string) MintPlugin {
	return MintPlugin{name: name}
}

func (mp MintPlugin) Name() string {
	return mp.name
}

// Set initial minters
func (mp MintPlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	// value is always a hex-encoded address
	addr, err := hex.DecodeString(value)
	if err != nil {
		return fmt.Sprintf("Invalid address: %s: %v", addr, err)
	}

	switch key {
	case AddBanker:
		s := mp.loadState(store)
		s.AddBanker(addr)
		mp.saveState(store, s)
		mp.saveState(store, s)
		return fmt.Sprintf("Added: %s", addr)
	case RemoveBanker:
		s := mp.loadState(store)
		s.RemoveBanker(addr)
		mp.saveState(store, s)
		return fmt.Sprintf("Removed: %s", addr)
	default:
		return fmt.Sprintf("Unknown key: %s", key)
	}
}

// This allows
func (mp MintPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
	// parse transaction
	var tx MintTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrEncodingError
	}

	// make sure it was signed by a banker
	s := mp.loadState(store)
	if !s.IsBanker(ctx.CallerAddress) {
		return abci.ErrUnauthorized
	}

	// now, send all this money!
	for _, winner := range tx.Winners {
		// load or create account
		acct := state.GetAccount(store, winner.Addr)
		if acct == nil {
			acct = &types.Account{
				PubKey:   nil,
				Sequence: 0,
			}
		}

		// add the money
		acct.Balance = acct.Balance.Plus(winner.Amount)

		// and save the new balance
		state.SetAccount(store, winner.Addr, acct)
	}

	return abci.Result{}
}

// placeholders empty to fulfill interface
func (mp MintPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {}
func (mp MintPlugin) BeginBlock(store types.KVStore, height uint64)         {}
func (mp MintPlugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	return nil
}

/*** implementation ***/

func (mp MintPlugin) stateKey() []byte {
	key := fmt.Sprintf("*%s*", mp.name)
	return []byte(key)
}

func (mp MintPlugin) loadState(store types.KVStore) *MintState {
	var s MintState
	data := store.Get(mp.stateKey())
	// here return an uninitialized state
	if len(data) == 0 {
		return &s
	}

	err := wire.ReadBinaryBytes(data, &s)
	// this should never happen, but we should also never panic....
	if err != nil {
		panic(err)
	}
	return &s
}

func (mp MintPlugin) saveState(store types.KVStore, state *MintState) {
	value := wire.BinaryBytes(*state)
	store.Set(mp.stateKey(), value)
}

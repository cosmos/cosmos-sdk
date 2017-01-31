package escrow

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
)

// EscrowPlugin is a plugin, storing all state prefixed with it's unique name
type EscrowPlugin struct {
	name string
}

func New(name string) EscrowPlugin {
	return EscrowPlugin{name: name}
}

func (mp EscrowPlugin) Name() string {
	return mp.name
}

// SetOption not supported by EscrowPlugin
func (mp EscrowPlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return fmt.Sprintf("Unknown key: %s", key)
}

// This allows
func (mp EscrowPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
	// // parse transaction
	// var tx MintTx
	// err := wire.ReadBinaryBytes(txBytes, &tx)
	// if err != nil {
	// 	return abci.ErrEncodingError
	// }

	// // make sure it was signed by a banker
	// s := mp.loadState(store)
	// if !s.IsBanker(ctx.CallerAddress) {
	// 	return abci.ErrUnauthorized
	// }

	// // now, send all this money!
	// for _, winner := range tx.Winners {
	// 	// load or create account
	// 	acct := state.GetAccount(store, winner.Addr)
	// 	if acct == nil {
	// 		acct = &types.Account{
	// 			PubKey:   nil,
	// 			Sequence: 0,
	// 		}
	// 	}

	// 	// add the money
	// 	acct.Balance = acct.Balance.Plus(winner.Amount)

	// 	// and save the new balance
	// 	state.SetAccount(store, winner.Addr, acct)
	// }

	return abci.Result{}
}

// placeholders empty to fulfill interface
func (mp EscrowPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {}
func (mp EscrowPlugin) BeginBlock(store types.KVStore, height uint64)         {}
func (mp EscrowPlugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	return nil
}

/*** implementation ***/

// func (mp EscrowPlugin) stateKey() []byte {
// 	key := fmt.Sprintf("*%s*", mp.name)
// 	return []byte(key)
// }

// func (mp EscrowPlugin) loadState(store types.KVStore) *MintState {
// 	var s MintState
// 	data := store.Get(mp.stateKey())
// 	// here return an uninitialized state
// 	if len(data) == 0 {
// 		return &s
// 	}

// 	err := wire.ReadBinaryBytes(data, &s)
// 	// this should never happen, but we should also never panic....
// 	if err != nil {
// 		panic(err)
// 	}
// 	return &s
// }

// func (mp EscrowPlugin) saveState(store types.KVStore, state *MintState) {
// 	value := wire.BinaryBytes(*state)
// 	store.Set(mp.stateKey(), value)
// }

package options

import (
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
)

// All concepts related to payments should go here
type Accountant struct {
	store types.KVStore
}

func (a Accountant) GetAccount(addr []byte) *types.Account {
	return state.GetAccount(a.store, addr)
}

func (a Accountant) SetAccount(addr []byte, acc *types.Account) {
	state.SetAccount(a.store, addr, acc)
}

func (a Accountant) Refund(ctx types.CallContext) {
	a.Pay(ctx.CallerAddress, ctx.Coins)
}

func (a Accountant) Pay(addr []byte, coins types.Coins) {
	acct := a.GetAccount(addr)
	if acct == nil {
		acct = &types.Account{}
	}
	acct.Balance = acct.Balance.Plus(coins)
	a.SetAccount(addr, acct)
}

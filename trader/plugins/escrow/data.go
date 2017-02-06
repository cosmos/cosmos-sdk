package escrow

import (
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
)

// Payback is used to signal who to send the money to
type Payback struct {
	Addr   []byte
	Amount types.Coins
}

func paybackCtx(ctx types.CallContext) Payback {
	return Payback{
		Addr:   ctx.CallerAddress,
		Amount: ctx.Coins,
	}
}

// Pay is used to return money back to one person after the transaction
// this could refund the fees, or pay out escrow, or anything else....
func (p Payback) Pay(store types.KVStore) {
	if len(p.Addr) == 20 {
		acct := state.GetAccount(store, p.Addr)
		if acct == nil {
			acct = &types.Account{}
		}
		acct.Balance = acct.Balance.Plus(p.Amount)
		state.SetAccount(store, p.Addr, acct)
	}
}

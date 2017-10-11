package coin

import (
	"testing"

	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

func makeHandler() stack.Dispatchable {
	return NewHandler()
}

func makeSimpleTx(from, to sdk.Actor, amount Coins) sdk.Tx {
	in := []TxInput{{Address: from, Coins: amount}}
	out := []TxOutput{{Address: to, Coins: amount}}
	return NewSendTx(in, out)
}

func BenchmarkSimpleTransfer(b *testing.B) {
	h := makeHandler()
	store := state.NewMemKVStore()
	logger := log.NewNopLogger()

	// set the initial account
	acct := NewAccountWithKey(Coins{{"mycoin", 1234567890}})
	h.InitState(logger, store, NameCoin, "account", acct.MakeOption(), nil)
	sender := acct.Actor()
	receiver := sdk.Actor{App: "foo", Address: cmn.RandBytes(20)}

	// now, loop...
	for i := 1; i <= b.N; i++ {
		ctx := stack.MockContext("foo", 100).WithPermissions(sender)
		tx := makeSimpleTx(sender, receiver, Coins{{"mycoin", 2}})
		_, err := h.DeliverTx(ctx, store, tx, nil)
		// never should error
		if err != nil {
			panic(err)
		}
	}

}

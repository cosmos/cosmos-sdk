package bank

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/mock"

	abci "github.com/tendermint/abci/types"
)

func BenchmarkOneBankSendTxPerBlock(b *testing.B) {
	benchmarkApp, _ := getBenchmarkMockApp()

	// Add an account at genesis
	acc := &auth.BaseAccount{
		Address: addr1,
		// Some value conceivably higher than the benchmarks would ever go
		Coins: sdk.Coins{sdk.NewCoin("foocoin", 100000000000)},
	}
	accs := []auth.Account{acc}

	// Construct genesis state
	mock.SetGenesis(benchmarkApp, accs)
	// Precompute all txs
	txs := mock.GenSequenceOfTxs([]sdk.Msg{sendMsg1}, []int64{0}, []int64{int64(0)}, b.N, priv1)
	b.ResetTimer()
	// Run this with a profiler, so its easy to distinguish what time comes from
	// Committing, and what time comes from Check/Deliver Tx.
	for i := 0; i < b.N; i++ {
		benchmarkApp.BeginBlock(abci.RequestBeginBlock{})
		x := benchmarkApp.Check(txs[i])
		if !x.IsOK() {
			panic("something is broken in checking transaction")
		}
		benchmarkApp.Deliver(txs[i])
		benchmarkApp.EndBlock(abci.RequestEndBlock{})
		benchmarkApp.Commit()
	}
}

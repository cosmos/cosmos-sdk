package benchmarking

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/mock"
	"github.com/cosmos/cosmos-sdk/x/bank"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

var (
	priv1    = crypto.GenPrivKeyEd25519()
	addr1    = priv1.PubKey().Address()
	priv2    = crypto.GenPrivKeyEd25519()
	addr2    = priv2.PubKey().Address()
	coins    = sdk.Coins{sdk.NewCoin("foocoin", 10)}
	sendMsg1 = bank.MsgSend{
		Inputs:  []bank.Input{bank.NewInput(addr1, coins)},
		Outputs: []bank.Output{bank.NewOutput(addr2, coins)},
	}
)

func BenchmarkOneBankSendTxPerBlock(b *testing.B) {
	benchmarkApp, _ := bank.GetBenchmarkMockApp()

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
	// Commiting, and what time comes from Check/Deliver Tx.
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

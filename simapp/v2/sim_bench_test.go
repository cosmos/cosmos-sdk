//go:build sims

package simapp

import (
	"github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"testing"
)

func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()
	cfg := cli.NewConfigFromFlags()
	cfg.ChainID = SimAppChainID
	for i := 0; i < b.N; i++ {
		RunWithSeed[Tx](b, NewSimApp[Tx], AppConfig, cfg, 1)
	}
}

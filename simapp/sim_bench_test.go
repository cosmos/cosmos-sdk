//go:build sims

package simapp

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutils/sims"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ cosmossdk.io/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = sims.SimAppChainID

	sims.RunWithSeed(b, config, NewSimApp, setupStateFactory, 1, nil)
}

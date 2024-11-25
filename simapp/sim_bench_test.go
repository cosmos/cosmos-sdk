//go:build sims

package simapp

import (
	"github.com/cosmos/cosmos-sdk/simsx"
	"testing"

	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ cosmossdk.io/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = simsx.SimAppChainID

	simsx.RunWithSeed(b, config, NewSimApp, setupStateFactory, 1, nil)
}

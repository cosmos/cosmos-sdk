//go:build sims

package simapp

import (
	"github.com/cosmos/cosmos-sdk/simsx"
	"testing"

<<<<<<< HEAD
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/testutils/sims"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
=======
>>>>>>> bf7768006 (feat(sims): Add sims2 framework and factory methods (#21613))
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

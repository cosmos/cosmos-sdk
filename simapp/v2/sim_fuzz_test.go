//go:build sims

package simapp

import (
	simsxv2 "github.com/cosmos/cosmos-sdk/simsx/v2"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"testing"
)

func FuzzFullAppSimulation(f *testing.F) {
	cfg := simcli.NewConfigFromFlags()
	cfg.ChainID = SimAppChainID

	f.Fuzz(func(t *testing.T, rawSeed []byte) {
		if len(rawSeed) < 8 {
			t.Skip()
			return
		}
		randSource := simsxv2.NewByteSource(cfg.FuzzSeed, cfg.Seed)
		RunWithRandSource[Tx](t, NewSimApp[Tx], AppConfig, cfg, randSource)
	})
}

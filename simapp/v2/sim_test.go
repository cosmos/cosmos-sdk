package simapp

import "testing"

func TestSimsAppV2(t *testing.T) {
	RunWithSeeds[Tx](t, defaultSeeds)
	// RunWithSeed(t, cli.NewConfigFromFlags(), 99)
}

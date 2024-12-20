package simapp

import (
	"cosmossdk.io/core/transaction"
	"testing"
)

func TestSimsAppV2(t *testing.T) {
	RunWithSeeds[transaction.Tx](t, NewSimApp[transaction.Tx], AppConfig, DefaultSeeds)
}

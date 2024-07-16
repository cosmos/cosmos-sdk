package systemtests

import (
	"testing"
	"time"
)

func TestWaitForBlock(t *testing.T) {
	// Scenario:
	// Spin up a chain and wait for height 5

	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)

	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)

	// query validator address to delegate tokens
	sut.AwaitBlockHeight(t, 12, time.Second*30)
}

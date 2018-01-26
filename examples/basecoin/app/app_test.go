package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
)

func TestSendMsg(t *testing.T) {
	tba := newTestBasecoinApp()
	tba.RunBeginBlock()

	// Construct a SendMsg.
	var msg = bank.SendMsg{
		Inputs: []bank.Input{
			{
				Address:  crypto.Address([]byte("input")),
				Coins:    sdk.Coins{{"atom", 10}},
				Sequence: 1,
			},
		},
		Outputs: []bank.Output{
			{
				Address: crypto.Address([]byte("output")),
				Coins:   sdk.Coins{{"atom", 10}},
			},
		},
	}

	// Run a Check on SendMsg.
	res := tba.RunCheckMsg(msg)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Run a Deliver on SendMsg.
	res = tba.RunDeliverMsg(msg)
	assert.Equal(t, sdk.CodeUnrecognizedAddress, res.Code, res.Log)
}

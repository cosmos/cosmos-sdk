package paytovote

import (
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/testutils"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	eyescli "github.com/tendermint/merkleeyes/client"
)

func TestP2VPlugin(t *testing.T) {

	// Basecoin initialization
	eyesCli := eyescli.NewLocalClient()
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(eyesCli)
	bcApp.SetOption("base/chainID", chainID)
	t.Log(bcApp.Info())

	// Add Counter plugin
	P2VPluginName := "testP2V"
	P2VPlugin := New(P2VPluginName)
	bcApp.RegisterPlugin(P2VPlugin)

	// Account initialization
	test1PrivAcc := testutils.PrivAccountFromSecret("test1")

	// Seed Basecoin with account
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"voteToken", 1000}, {"issueToken", 1000}}
	bcApp.SetOption("base/account", string(wire.JSONBytes(test1Acc)))

	DeliverTx := func(gas int64,
		fee types.Coin,
		inputCoins types.Coins,
		inputSequence int,
		issue string,
		actionTypeByte byte,
		cost2Vote,
		cost2CreateIssue types.Coins) abci.Result {

		// Construct an AppTx signature
		tx := &types.AppTx{
			Gas:   gas,
			Fee:   fee,
			Name:  P2VPluginName,
			Input: types.NewTxInput(test1Acc.PubKey, inputCoins, inputSequence),
			Data: wire.BinaryBytes(
				P2VTx{
					Valid:            true,
					Issue:            issue,
					ActionTypeByte:   actionTypeByte,
					Cost2Vote:        cost2Vote,
					Cost2CreateIssue: cost2CreateIssue,
				}),
		}

		// Sign request
		signBytes := tx.SignBytes(chainID)
		t.Logf("Sign bytes: %X\n", signBytes)
		sig := test1PrivAcc.PrivKey.Sign(signBytes)
		tx.Input.Signature = sig
		t.Logf("Signed TX bytes: %X\n", wire.BinaryBytes(struct{ types.Tx }{tx}))

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		return bcApp.DeliverTx(txBytes)
	}

	// REF: DeliverCounterTx(gas, fee, inputCoins, inputSequence, issue, action, cost2Vote, cost2CreateIssue)

	// Test a basic send, no fee
	res := DeliverTx(0, types.Coin{}, types.Coins{{"voteToken", 1}, {"issueToken", 1}}, 1,
		"free internet", TypeByteCreateIssue, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 1}})
	assert.True(t, res.IsOK(), res.String())

	/*
		// Test fee prevented transaction
		res = DeliverCounterTx(0, types.Coin{"", 2}, types.Coins{{"", 1}}, 2, types.Coins{})
		assert.True(t, res.IsErr(), res.String())

		// Test input equals fee
		res = DeliverCounterTx(0, types.Coin{"", 2}, types.Coins{{"", 2}}, 2, types.Coins{})
		assert.True(t, res.IsOK(), res.String())

		// Test more input than fee
		res = DeliverCounterTx(0, types.Coin{"", 2}, types.Coins{{"", 3}}, 3, types.Coins{})
		assert.True(t, res.IsOK(), res.String())

		// Test input equals fee+cost
		res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 3}, {"gold", 1}}, 4, types.Coins{{"", 2}, {"gold", 1}})
		assert.True(t, res.IsOK(), res.String())

		// Test fee+cost prevented transaction, not enough ""
		res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 2}, {"gold", 1}}, 5, types.Coins{{"", 2}, {"gold", 1}})
		assert.True(t, res.IsErr(), res.String())

		// Test fee+cost prevented transaction, not enough "gold"
		res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 3}, {"gold", 1}}, 5, types.Coins{{"", 2}, {"gold", 2}})
		assert.True(t, res.IsErr(), res.String())

		// Test more input than fee, more ""
		res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 4}, {"gold", 1}}, 6, types.Coins{{"", 2}, {"gold", 1}})
		assert.True(t, res.IsOK(), res.String())

		// Test more input than fee, more "gold"
		res = DeliverCounterTx(0, types.Coin{"", 1}, types.Coins{{"", 3}, {"gold", 2}}, 7, types.Coins{{"", 2}, {"gold", 1}})
		assert.True(t, res.IsOK(), res.String())
	*/
}

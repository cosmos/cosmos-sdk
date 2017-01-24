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
	test1Acc.Balance = types.Coins{{"", 1000}, {"issueToken", 1000}, {"voteToken", 1000}}
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

	//TODO: Generate tests which  query the results of an issue
	/*	queryIssue := func(issue string) abci.Result {
		key := P2VPlugin.StateKey(issue)
		query := make([]byte, 1+wire.ByteSliceSize(key))
		buf := query
		buf[0] = 0x01 // Get TypeByte
		buf = buf[1:]
		wire.PutByteSlice(buf, key)
		t.Log(len(query))
		return bcApp.Query(query)
	}*/
	// REF: DeliverCounterTx(gas, fee, inputCoins, inputSequence, issue, action, cost2Vote, cost2CreateIssue)

	issue1 := "free internet"
	issue2 := "commutate foobar"

	// Test a basic issue generation
	res := DeliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 1,
		issue1, TypeByteCreateIssue, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 1}})
	assert.True(t, res.IsOK(), res.String())

	// Test a basic votes
	res = DeliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 2,
		issue1, TypeByteVoteFor, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 1}})
	assert.True(t, res.IsOK(), res.String())

	res = DeliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 3,
		issue1, TypeByteVoteAgainst, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 1}})
	assert.True(t, res.IsOK(), res.String())

	res = DeliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 4,
		issue1, TypeByteVoteSpoiled, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 1}})
	assert.True(t, res.IsOK(), res.String())

	// Test prevented voting on non-existent issue
	res = DeliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 5,
		issue2, TypeByteVoteFor, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 1}})
	assert.True(t, res.IsErr(), res.String())

	// Test prevented duplicate issue generation
	res = DeliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 5,
		issue1, TypeByteCreateIssue, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 1}})
	assert.True(t, res.IsErr(), res.String())

	// Test prevented issue generation from insufficient funds
	res = DeliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 5,
		issue2, TypeByteCreateIssue, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 2}})
	assert.True(t, res.IsErr(), res.String())

	// Test prevented voting from insufficient funds
	res = DeliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 5,
		issue1, TypeByteVoteFor, types.Coins{{"voteToken", 2}}, types.Coins{{"issueToken", 1}})
	assert.True(t, res.IsErr(), res.String())
}

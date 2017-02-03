package paytovote

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/basecoin/app"
	//cmds "github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/testutils"
	"github.com/tendermint/basecoin/types"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	eyescli "github.com/tendermint/merkleeyes/client"
)

func TestP2VPlugin(t *testing.T) {

	// Basecoin initialization
	store := eyescli.NewLocalClient("", 0) //non-persistent instance of merkleeyes
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(store)
	bcApp.SetOption("base/chainID", chainID)

	// Add Counter plugin
	P2VPlugin := New()
	bcApp.RegisterPlugin(P2VPlugin)

	// Account initialization
	test1PrivAcc := testutils.PrivAccountFromSecret("test1")
	test1Acc := test1PrivAcc.Account

	// Seed Basecoin with account
	startBal := types.Coins{{"", 1000}, {"issueToken", 1000}, {"voteToken", 1000}}
	test1Acc.Balance = startBal
	bcApp.SetOption("base/account", string(wire.JSONBytes(test1Acc)))
	bcApp.Commit()

	deliverTx := func(gas int64,
		fee types.Coin,
		inputCoins types.Coins,
		inputSequence int,
		txData []byte) abci.Result {

		// Construct an AppTx signature
		tx := &types.AppTx{
			Gas:   gas,
			Fee:   fee,
			Name:  P2VPlugin.Name(),
			Input: types.NewTxInput(test1Acc.PubKey, inputCoins, inputSequence),
			Data:  txData,
		}

		// Sign request
		signBytes := tx.SignBytes(chainID)
		sig := test1PrivAcc.PrivKey.Sign(signBytes)
		tx.Input.Signature = sig

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		return bcApp.DeliverTx(txBytes)
	}

	testBalance := func(expected types.Coins) {

		acc := state.GetAccount(bcApp.GetState(), test1Acc.PubKey.Address())
		if acc == nil {
			panic("nil account when trying compare balance")
		}

		bal := acc.Balance
		if !bal.IsEqual(expected) {
			var expStr, balStr string
			for i := 0; i < len(expected); i++ {
				expStr += " " + expected[i].String()
			}
			for i := 0; i < len(bal); i++ {
				balStr += " " + bal[i].String()
			}

			panic(cmn.Fmt("bad balance expected %v, got %v", expStr, balStr))
		}
	}

	//test for an issue that shouldn't exist
	testNoIssue := func(issue string) {
		_, err := getIssue(bcApp.GetState(), issue)
		if err == nil {
			panic(cmn.Fmt("issue that shouldn't exist was found, issue: %v", issue))
		}
	}

	//test for an issue that should exist
	testIssue := func(issue string, expFor, expAgainst int) {
		p2vIssue, err := getIssue(bcApp.GetState(), issue)

		// return //TODO fix these tests, bad store being accessed

		//test for errors
		if err != nil {
			panic(cmn.Fmt("error loading issue %v for issue test, error: %v", issue, err.Error()))
		}

		if p2vIssue.VotesFor != expFor {
			panic(cmn.Fmt("expected %v votes-for, got %v votes-for, for issue %v", expFor, p2vIssue.VotesFor, issue))
		}

		if p2vIssue.VotesAgainst != expAgainst {
			panic(cmn.Fmt("expected %v votes-against, got %v votes-against, for issue %v", expAgainst, p2vIssue.VotesAgainst, issue))
		}
	}

	// REF: deliverTx(gas, fee, inputCoins, inputSequence, NewVoteTxBytes(issue, voteTypeByte))
	// REF: deliverTx(gas, fee, inputCoins, inputSequence, NewCreateIssueTxBytes(issue, feePerVote, fee2CreateIssue))

	issue1 := "free internet"
	issue2 := "commutate foobar"

	// Test a basic issue generation
	res := deliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 2}}, 1,
		NewCreateIssueTxBytes(issue1, types.Coins{{"voteToken", 2}}, types.Coins{{"issueToken", 1}}))
	assert.True(t, res.IsOK(), res.String())
	testBalance(startBal.Minus(types.Coins{{"issueToken", 1}}))
	testIssue(issue1, 0, 0)

	// Test a basic votes
	res = deliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 2}}, 2,
		NewVoteTxBytes(issue1, TypeByteVoteFor))
	assert.True(t, res.IsOK(), res.String())
	testBalance(startBal.Minus(types.Coins{{"issueToken", 1}, {"voteToken", 2}}))
	testIssue(issue1, 1, 0)

	res = deliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 2}}, 3,
		NewVoteTxBytes(issue1, TypeByteVoteAgainst))
	assert.True(t, res.IsOK(), res.String())
	testBalance(startBal.Minus(types.Coins{{"issueToken", 1}, {"voteToken", 4}}))
	testIssue(issue1, 1, 1)

	// Test prevented voting on non-existent issue
	res = deliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 2}}, 4,
		NewVoteTxBytes(issue2, TypeByteVoteFor))
	assert.True(t, res.IsErr(), res.String())
	testBalance(startBal.Minus(types.Coins{{"issueToken", 1}, {"voteToken", 4}}))
	testNoIssue(issue2)

	// Test prevented duplicate issue generation
	res = deliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 2}}, 5,
		NewCreateIssueTxBytes(issue1, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 1}}))
	assert.True(t, res.IsErr(), res.String())
	testBalance(startBal.Minus(types.Coins{{"issueToken", 1}, {"voteToken", 4}}))

	// Test prevented issue generation from insufficient funds
	res = deliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 2}}, 6,
		NewCreateIssueTxBytes(issue2, types.Coins{{"voteToken", 1}}, types.Coins{{"issueToken", 2}}))
	assert.True(t, res.IsErr(), res.String())
	testBalance(startBal.Minus(types.Coins{{"issueToken", 1}, {"voteToken", 4}}))
	testNoIssue(issue2)

	// Test prevented voting from insufficient funds
	res = deliverTx(0, types.Coin{}, types.Coins{{"", 1}, {"issueToken", 1}, {"voteToken", 1}}, 7,
		NewVoteTxBytes(issue1, TypeByteVoteFor))
	assert.True(t, res.IsErr(), res.String())
	testBalance(startBal.Minus(types.Coins{{"issueToken", 1}, {"voteToken", 4}}))
	testIssue(issue1, 1, 1)
}

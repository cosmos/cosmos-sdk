package simpleGovernance

import (
	// 	"os"

	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSubmitProposalMsg(t *testing.T) {

	//TODO Test:
	// Good proposal --> OK
	// deposit greater than account balance --> error

	// crete context and proposalKeeper each address has 200 Atoms
	ctx, _, k := createTestInput(t, true, int64(200))

	err := checkProposal(ctx, k) // No proposal
	assert.NotNil(t, err)

	cases := []struct {
		valid bool
		msg   SubmitProposalMsg
	}{
		{true, NewSubmitProposalMsg(titles[0], descriptions[0], ctx.BlockHeight(), coinsHandlerTest[0], addrs[0])},
		{false, NewSubmitProposalMsg(titles[1], descriptions[1], ctx.BlockHeight(), coinsHandlerTest[1], addrs[1])}, // empty coins
		{false, NewSubmitProposalMsg(titles[2], descriptions[2], ctx.BlockHeight(), coinsHandlerTest[2], addrs[2])}, // balance below deposit
	}

	for i := range cases {
		res := handleSubmitProposalMsg(ctx, k, cases[i].msg)
		if cases[i].valid {
			fmt.Println(res.Tags)
			kvPair := res.Tags.ToKVPairs()
			fmt.Println(kvPair)
			assert.True(t, res.IsOK(), "%d: %+v", i, res)
		} else {
			assert.False(t, res.IsOK(), "%d", i)
		}
	}

	// Test if the proposal reached the end of voting period
	err = checkProposal(ctx, k)
	assert.Nil(t, err)

	// Voting Handler

	// TODO get proposalID from handleSubmitProposalMsg response

	//TODO Test:
	// Valid Msg ->
	// Invalid ProposalID
	// Proposal is not open
	// No delegations for the address

	validatorBond := sdk.Coin{"Atom", 80}
	declareDescription := stake.NewDescription("moniker", "identity", "website", "details")
	declareMsg := stake.NewMsgCreateValidator(addrs[0], pks[0], validatorBond, declareDescription)

	delegateBond := sdk.Coin{"Atom", 20}
	delegateMsg := stake.NewMsgDelegate(addrs[1], addrs[0], delegateBond)

	// TODO handle delegation

	voteCases := []struct {
		valid bool
		msg   VoteMsg
	}{
		{true, NewVoteMsg(1, options[0], addrs[1])},
		{false, NewVoteMsg(2, options[1], addrs[1])}, // invalid proposalID
		{false, NewVoteMsg(1, options[1], addrs[2])}, // voter has no delegations
	}

	for i := range cases {
		res := handleVoteMsg(ctx, k, voteCases[i].msg)
		if voteCases[i].valid {
			fmt.Println(res.Tags)
			kvPair := res.Tags.ToKVPairs()
			fmt.Println(kvPair)
			assert.True(t, res.IsOK(), "%d: %+v", i, res)
		} else {
			assert.False(t, res.IsOK(), "%d", i)
		}
	}

	// Proposal is not open
	proposal, err := k.GetProposal(ctx, 1)
	require.NotNil(t, err)
	proposal.State = "Accepted"
	err = k.SetProposal(ctx, 1, proposal)
	require.NotNil(t, err)
	msg := NewVoteMsg(1, options[1], addrs[6])
	res := handleVoteMsg(ctx, k, msg)
	assert.False(t, res.IsOK())

	proposal.State = "Open"

	err = k.SetProposal(ctx, 1, proposal)

	require.NotNil(t, err)

}

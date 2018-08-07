package simpleGovernance

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestSimpleGovKeeper(t *testing.T) {

	mapp, k, _, _, _, _ := getMockApp(t, 100)

	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	if ctx.KVStore(k.storeKey) == nil {
		panic("Nil interface")
	}
	// newTags := sdk.NewTags()
	// _, err := checkProposal(ctx, k, newTags) // No proposal
	// assert.NotNil(t, err)

	// create proposals
	proposal := k.NewProposal(ctx, titles[0], descriptions[0])
	// TODO Error when defined proposal 2 is defined here
	// proposal2 := k.NewProposal(ctx, titles[1], descriptions[1])

	// –––––––––––––––––––––––––––––––––––––––
	//                KEEPER
	// –––––––––––––––––––––––––––––––––––––––

	// ––––––– Test GetProposal –––––––

	// Case 1: valid request
	res, err := k.GetProposal(ctx, proposal.ID)
	assert.NotEqual(t, Proposal{}, res)
	assert.Nil(t, err)
	assert.Equal(t, proposal, res)

	proposal2 := k.NewProposal(ctx, titles[1], descriptions[1])

	// Case 2: invalid proposalID
	res, err = k.GetProposal(ctx, 4)
	assert.NotNil(t, err)

	k.SetVote(ctx, 1, addrs[2], options[1])

	// ––––––– Test GetVote –––––––

	// Case 1: existing proposal, valid voter
	option, err := k.GetVote(ctx, 1, addrs[2])
	assert.Equal(t, options[1], option)
	assert.Nil(t, err)

	// Case 2: existing proposal, invalid voter
	option, err = k.GetVote(ctx, 1, addrs[3]) // existing proposal, invalid voter
	assert.Equal(t, "", option)
	assert.NotNil(t, err)

	// Case 3: invalid proposal, valid voter
	option, err = k.GetVote(ctx, 2, addrs[2])
	assert.Equal(t, "", option)
	assert.NotNil(t, err)

	// Case 4: invalid proposal, invalid voter
	option, err = k.GetVote(ctx, 2, addrs[3])
	assert.Equal(t, "", option)
	assert.NotNil(t, err)

	// –––––––––––––––––––––––––––––––––––––––
	//             KEEPER READ
	// –––––––––––––––––––––––––––––––––––––––

	simpleGovKey := sdk.NewKVStoreKey("simpleGov")
	keeperRead := NewKeeperRead(simpleGovKey, k.ck, k.sm, DefaultCodespace)

	// ––––––– Test GetProposal –––––––

	// Case 1: valid request
	res, err = keeperRead.GetProposal(ctx, proposal.ID)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, proposal, res)

	// Case 2: invalid proposalID
	res, err = keeperRead.GetProposal(ctx, 10)
	assert.NotNil(t, err)

	// ––––––– Test SetProposal –––––––

	err = keeperRead.SetProposal(ctx, proposal2) // Error Unauthorized
	assert.NotNil(t, err)

	// ––––––– Test GetVote –––––––

	// Case 1: existing proposal, valid voter
	option, err = keeperRead.GetVote(ctx, 1, addrs[2])
	assert.Equal(t, options[1], option)
	assert.Nil(t, err)

	// Case 2: existing proposal, invalid voter
	option, err = keeperRead.GetVote(ctx, 1, addrs[3]) // existing proposal, invalid voter
	assert.Equal(t, "", option)
	assert.NotNil(t, err)

	// Case 3: invalid proposal, valid voter
	option, err = keeperRead.GetVote(ctx, 2, addrs[2])
	assert.Equal(t, "", option)
	assert.NotNil(t, err)

	// Case 4: invalid proposal, invalid voter
	option, err = keeperRead.GetVote(ctx, 2, addrs[3])
	assert.Equal(t, "", option)
	assert.NotNil(t, err)

	// –––––––––––––––––––––––––––––––––––––––
	//            PROPOSAL QUEUE
	// –––––––––––––––––––––––––––––––––––––––

	err = k.ProposalQueuePush(ctx, 1) // ProposalQueue not set
	assert.NotNil(t, err)

	k.setProposalQueue(ctx, ProposalQueue{})

	err = k.ProposalQueuePush(ctx, 1)
	assert.Nil(t, err)

	res, err = k.ProposalQueueHead(ctx) // Gets first proposal
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, proposal, res)

	res, err = k.ProposalQueuePop(ctx) // Pops first proposal
	assert.Nil(t, err)
	assert.Equal(t, proposal, res)

	res, err = k.ProposalQueuePop(ctx) // Empty queue --> error
	assert.NotNil(t, err)
	assert.Equal(t, proposal, Proposal{})

	res, err = k.ProposalQueueHead(ctx) // Empty queue --> error
	assert.NotNil(t, err)
}

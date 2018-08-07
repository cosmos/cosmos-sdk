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
	if ctx.KVStore(k.SimpleGov) == nil {
		panic("Nil interface")
	}
	// newTags := sdk.NewTags()
	// _, err := checkProposal(ctx, k, newTags) // No proposal
	// assert.NotNil(t, err)

	// create proposals
	proposal := k.NewProposal(ctx, titles[1], descriptions[1])
	proposal2 := k.NewProposal(ctx, titles[2], descriptions[2])

	// –––––––––––––––––––––––––––––––––––––––
	//                KEEPER
	// –––––––––––––––––––––––––––––––––––––––

	// ––––––– Test GetProposal –––––––

	// Case 1: valid request
	resProposal, err := k.GetProposal(ctx, 1)
	assert.NotNil(t, resProposal)
	assert.Nil(t, err)
	assert.Equal(t, proposal, resProposal)

	// Case 2: invalid proposalID
	resProposal, err = k.GetProposal(ctx, 2)
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
	resProposal, err = keeperRead.GetProposal(ctx, 1)
	assert.NotNil(t, resProposal)
	assert.Nil(t, err)
	assert.Equal(t, proposal, resProposal)

	// Case 2: invalid proposalID
	resProposal, err = keeperRead.GetProposal(ctx, 2)
	assert.NotNil(t, err)

	// ––––––– Test SetProposal –––––––

	err = keeperRead.SetProposal(ctx, 2, proposal2) // Error Unauthorized
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

	resProposal, err = k.ProposalQueueHead(ctx) // Gets first proposal
	assert.NotNil(t, resProposal)
	assert.Nil(t, err)
	assert.Equal(t, proposal, resProposal)

	resProposal, err = k.ProposalQueuePop(ctx) // Pops first proposal
	assert.Nil(t, err)
	assert.Equal(t, proposal, resProposal)

	resProposal, err = k.ProposalQueuePop(ctx) // Empty queue --> error
	assert.NotNil(t, err)
	assert.Equal(t, proposal, Proposal{})

	resProposal, err = k.ProposalQueueHead(ctx) // Empty queue --> error
	assert.NotNil(t, err)
}

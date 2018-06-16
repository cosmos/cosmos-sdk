package simpleGovernance

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestSimpleGovKeeper(t *testing.T) {

	ctx, _, k := createTestInput(t, int64(200))
	assert.NotNil(t, k)

	// create proposals
	proposal := NewProposal(titles[1], description[1], addrs[1], ctx.BlockHeight(), sdk.Coins{{"Atom": 200}})
	proposal2 := NewProposal(titles[2], description[2], addrs[4], ctx.BlockHeight(), sdk.Coins{{"Atom": 150}})

	// –––––––––––––––––––––––––––––––––––––––
	//                KEEPER
	// –––––––––––––––––––––––––––––––––––––––

	// ––––––– Test SetProposal –––––––
	err := k.SetProposal(ctx, 1, proposal)
	assert.Nil(t, err)

	// ––––––– Test GetProposal –––––––

	// Case 1: valid request
	resProposal, err := k.GetProposal(ctx, 1)
	assert.NotNil(t, resProposal)
	assert.Nil(t, err)
	assert.Equal(t, proposal, resProposal)

	// Case 2: invalid proposalID
	resProposal, err := k.GetProposal(ctx, 2)
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
	option, err := k.GetVote(ctx, 2, addrs[2])
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
	resProposal, err := keeperRead.GetProposal(ctx, 1)
	assert.NotNil(t, resProposal)
	assert.Nil(t, err)
	assert.Equal(t, proposal, resProposal)

	// Case 2: invalid proposalID
	resProposal, err := keeperRead.GetProposal(ctx, 2)
	assert.NotNil(t, err)

	// ––––––– Test SetProposal –––––––

	err = keeperRead.SetProposal(ctx, 2, proposal2) // Error Unauthorized
	assert.NotNil(t, err)

	// ––––––– Test GetVote –––––––

	// Case 1: existing proposal, valid voter
	option, err := keeperRead.GetVote(ctx, 1, addrs[2])
	assert.Equal(t, options[1], option)
	assert.Nil(t, err)

	// Case 2: existing proposal, invalid voter
	option, err = keeperRead.GetVote(ctx, 1, addrs[3]) // existing proposal, invalid voter
	assert.Equal(t, "", option)
	assert.NotNil(t, err)

	// Case 3: invalid proposal, valid voter
	option, err := keeperRead.GetVote(ctx, 2, addrs[2])
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

	err = k.ProposalQueuePop(ctx) // Pops first proposal
	assert.Nil(t, err)

	err = k.ProposalQueuePop(ctx) // Empty queue --> error
	assert.NotNil(t, err)

	resProposal, err = k.ProposalQueueHead(ctx) // Empty queue --> error
	assert.NotNil(t, err)
}

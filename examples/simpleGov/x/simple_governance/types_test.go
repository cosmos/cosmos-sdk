package simpleGovernance

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestNewSubmitProposalMsg(t *testing.T) {
	goodTitle := "Photons at launch"
	emptyStr := ""
	goodDescription := "Should we include Photons at launch?"
	addr1 := sdk.Address([]byte{1, 2})
	addr2 := sdk.Address([]byte{7, 8})
	someCoins := sdk.Coins{{"atom", sdk.NewInt(int64(123))}}
	multiCoins := sdk.Coins{{"atom", sdk.NewInt(int64(123))}, {"eth", sdk.NewInt(int64(20))}}

	emptyAddr := sdk.Address{}
	emptyCoins := sdk.Coins{}
	emptyCoins2 := sdk.Coins{sdk.Coin{"eth", sdk.NewInt(int64(0))}}
	someEmptyCoins := sdk.Coins{{"eth", sdk.NewInt(int64(10))}, {"atom", sdk.NewInt(int64(0))}}
	minusCoins := sdk.Coins{{"eth", sdk.NewInt(int64(-34))}}
	someMinusCoins := sdk.Coins{{"atom", sdk.NewInt(int64(20))}, {"eth", sdk.NewInt(int64(-34))}}

	cases := []struct {
		valid bool
		spMsg SubmitProposalMsg
	}{
		{true, NewSubmitProposalMsg(goodTitle, goodDescription, someCoins, addr1)},
		{true, NewSubmitProposalMsg(goodTitle, goodDescription, multiCoins, addr2)},

		{false, NewSubmitProposalMsg(emptyStr, goodDescription, multiCoins, addr1)},          // empty title
		{false, NewSubmitProposalMsg(goodTitle, emptyStr, multiCoins, addr2)},                // empty description
		{false, NewSubmitProposalMsg(emptyStr, emptyStr, someCoins, addr2)},                  // invalid title and description
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, emptyCoins, addr1)},         // empty coins
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, emptyCoins2, addr1)},        // zero balance coins
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, someEmptyCoins, addr1)},     // zero balance coin
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, minusCoins, emptyAddr)},     // negative coin
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, someMinusCoins, emptyAddr)}, // negative coins
		{false, NewSubmitProposalMsg(emptyStr, emptyStr, someEmptyCoins, emptyAddr)},         // empty everything
	}

	for i, msg := range cases {
		err := msg.spMsg.ValidateBasic()
		if msg.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
			// GetSigners
			assert.Len(t, msg.spMsg.GetSigners(), 1)
			assert.Equal(t, msg.spMsg.GetSigners()[0], msg.spMsg.Submitter)
			// GetSignBytes
			assert.NotPanics(t, assert.PanicTestFunc(func() {
				msg.spMsg.GetSignBytes()
			}))
			assert.NotNil(t, msg.spMsg.GetSignBytes())
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestNewVoteMsg(t *testing.T) {
	emptyStr := ""
	addr1 := sdk.Address([]byte{1, 2})
	addr2 := sdk.Address([]byte{7, 8})

	emptyAddr := sdk.Address{}

	yay := "Yes"
	nay := "No"
	abstain := "Abstain"
	other := "Other"
	blank := "     "

	var posPropID int64 = 3
	var negPropID int64 = -8

	cases := []struct {
		valid   bool
		voteMsg VoteMsg
	}{
		{true, NewVoteMsg(posPropID, yay, addr1)},
		{true, NewVoteMsg(posPropID, nay, addr2)},
		{true, NewVoteMsg(posPropID, abstain, addr1)},
		{true, NewVoteMsg(posPropID, emptyStr, addr1)}, // empty Option is an Abstain

		{false, NewVoteMsg(negPropID, yay, addr1)},          // negative ProposalID
		{false, NewVoteMsg(posPropID, other, addr1)},        // other Option
		{false, NewVoteMsg(posPropID, blank, addr1)},        // blank Option
		{false, NewVoteMsg(posPropID, nay, emptyAddr)},      // empty Address
		{false, NewVoteMsg(negPropID, other, emptyAddr)},    // all wrong
		{false, NewVoteMsg(negPropID, emptyStr, emptyAddr)}, // all wrong
	}

	for i, msg := range cases {
		err := msg.voteMsg.ValidateBasic()
		if msg.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
			// GetSigners
			assert.Len(t, msg.voteMsg.GetSigners(), 1)
			assert.Equal(t, msg.voteMsg.GetSigners()[0], msg.voteMsg.Voter)
			// GetSignBytes
			assert.NotPanics(t, assert.PanicTestFunc(func() {
				msg.voteMsg.GetSignBytes()
			}))
			assert.NotNil(t, msg.voteMsg.GetSignBytes())
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestUpdateTally(t *testing.T) {

	goodTitle := "Photons at launch"
	goodDescription := "Should we include Photons at launch?"
	addr1 := sdk.Address([]byte{1, 2})
	someCoins := sdk.Coins{{"atom", sdk.NewInt(int64(123))}}

	proposal := NewProposal(goodTitle, goodDescription, addr1, 10, someCoins)
	proposalPtr := &proposal
	assert.Equal(t, int64(0), proposalPtr.YesVotes)
	assert.Equal(t, int64(0), proposalPtr.NoVotes)
	assert.Equal(t, int64(0), proposalPtr.AbstainVotes)

	// update YesVotes
	err := proposalPtr.updateTally("Yes", 20)
	assert.Nil(t, err)
	assert.Equal(t, int64(20), proposalPtr.YesVotes)
	// update NoVotes
	err = proposalPtr.updateTally("No", 8)
	assert.Nil(t, err)
	assert.Equal(t, int64(8), proposalPtr.NoVotes)
	// update AbstainVotes
	err = proposalPtr.updateTally("Abstain", 10)
	assert.Nil(t, err)
	assert.Equal(t, int64(10), proposalPtr.AbstainVotes)
	// invalid Option
	err = proposalPtr.updateTally("Whatever", 10)
	assert.NotNil(t, err)
	// blank Option
	err = proposalPtr.updateTally("     ", 10)
	assert.NotNil(t, err)
}

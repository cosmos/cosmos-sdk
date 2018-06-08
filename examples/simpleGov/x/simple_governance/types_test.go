package simpleGovernance

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
)

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.Address) {
	key := crypto.GenPrivKeyEd25519()
	pub := key.PubKey()
	addr := sdk.Address(pub.Address())
	return key, pub, addr
}

func TestNewSubmitProposalMsg(t *testing.T) {
	goodTitle := "Photons at launch"
	emptyStr := ""
	goodDescription := "Should we include Photons at launch?"
	addr1 := sdk.Address([]byte{1, 2})
	addr2 := sdk.Address([]byte{7, 8})
	someCoins := sdk.Coins{{"atom", 123}}
	multiCoins := sdk.Coins{{"atom", 123}, {"eth", 20}}
	var votingWindow1 int64 = 15

	emptyAddr := sdk.Address{}
	var zeroVotingWindow int64 = 0
	var negVotingWindow int64 = -20
	emptyCoins := sdk.Coins{}
	emptyCoins2 := sdk.Coins{sdk.Coin{"eth", 0}}
	someEmptyCoins := sdk.Coins{{"eth", 10}, {"atom", 0}}
	minusCoins := sdk.Coins{{"eth", -34}}
	someMinusCoins := sdk.Coins{{"atom", 20}, {"eth", -34}}

	cases := []struct {
		valid bool
		spMsg SubmitProposalMsg
	}{
		{true, NewSubmitProposalMsg(goodTitle, goodDescription, votingWindow1, someCoins, addr1)},
		{true, NewSubmitProposalMsg(goodTitle, goodDescription, votingWindow1, multiCoins, addr2)},

		{false, NewSubmitProposalMsg(emptyStr, goodDescription, votingWindow1, multiCoins, addr1)},             // empty title
		{false, NewSubmitProposalMsg(goodTitle, emptyStr, votingWindow1, multiCoins, addr2)},                   // empty description
		{false, NewSubmitProposalMsg(emptyStr, emptyStr, votingWindow1, someCoins, addr2)},                     // invalid title and description
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, votingWindow1, emptyCoins, addr1)},            // empty coins
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, votingWindow1, emptyCoins2, addr1)},           // zero balance coins
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, votingWindow1, someEmptyCoins, addr1)},        // zero balance coin
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, negVotingWindow, emptyCoins2, addr1)},         // negative votingWindow
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, zeroVotingWindow, someMinusCoins, emptyAddr)}, // zero VotingWindow
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, votingWindow1, minusCoins, emptyAddr)},        // negative coin
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, votingWindow1, someMinusCoins, emptyAddr)},    // negative coins
		{false, NewSubmitProposalMsg(emptyStr, emptyStr, zeroVotingWindow, someEmptyCoins, emptyAddr)},         // empty everything
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
	someCoins := sdk.Coins{{"atom", 123}}
	var votingWindow1 int64 = 15

	proposal := NewProposal(goodTitle, goodDescription, addr1, 10, votingWindow1, someCoins)
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
}

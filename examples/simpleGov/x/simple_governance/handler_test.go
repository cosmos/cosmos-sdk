package simpleGovernance

import (
	// 	"os"

	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestHandleSubmitProposalMsg(t *testing.T) {
	//TODO bad coins, bad address

	title := "Photons at launch"
	description := "Should we include Photons at launch?"
	addr1 := sdk.Address([]byte{1, 2})
	multiCoins := sdk.Coins{{"atom", 123}, {"eth", 20}}

	msg := NewSubmitProposalMsg(title, description, int64(20), multiCoins, addr1)

	// res := handleSubmitProposalMsg(ctx, k, msg)

	// tags := sdk.NewTags(
	// 	"action", []byte("propose"),
	// 	"proposal", int64ToBytes(proposalID),
	// 	"submitter", submitter.Bytes(),
	// )
	// kvpairs := res.Tags.ToKVPairs()
	// fmt.Println(kvpairs)

	// assert.Equal(t, , res)
}

func TestCheckProposal(t *testing.T) {
	//TODO
}

func TestHandleVoteMsg(t *testing.T) {
	//TODO
	// Proposal is not open
	// BlockHeight > Limit
	// Invalid ProposalID

	// No delegators ? for the address
	// voter option not found
	// invalid Option value
	// msg := NewVoteMsg(proposalID, option, voter)
	// res := handleVoteMsg(ctx, k, msg)
	// assert.Equal(t, sdk.Result{}, res)
}

func TestNewHandler(t *testing.T) {
	//TODO
	// create keeper, context and msgs
	// msg submmit proposal
	// msg vote
	// another msg
}

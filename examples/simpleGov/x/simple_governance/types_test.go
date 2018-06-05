package simpleGovernance

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
)

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"testing"
//
// 	"github.com/stretchr/testify/require"
// )

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.Address) {
	key := crypto.GenPrivKeyEd25519()
	pub := key.PubKey()
	addr := pub.Address()
	return key, pub, addr
}

func TestNewSubmitProposalMsg(t *testing.T) {
	goodTitle := "Photons at launch"
	var emptyStr = ""
	goodDescription := "Should we include Photons at launch?"
	addr1 := sdk.Address([]byte{1, 2})
	addr2 := sdk.Address([]byte{7, 8})
	someCoins := sdk.Coins{{"atom", 123}}
	multiCoins := sdk.Coins{{"atom", 123}, {"eth", 20}}

	var emptyAddr sdk.Address
	emptyCoins := sdk.Coins{}
	emptyCoins2 := sdk.Coins{{"eth", 0}}
	someEmptyCoins := sdk.Coins{{"eth", 10}, {"atom", 0}}
	minusCoins := sdk.Coins{{"eth", -34}}
	someMinusCoins := sdk.Coins{{"atom", 20}, {"eth", -34}}

	cases := []struct {
		valid bool
		spMsg SubmitProposalMsg
	}{
		{true, NewSubmitProposalMsg(goodTitle, goodDescription, someCoins, addr1)},
		{true, NewSubmitProposalMsg(goodTitle, goodDescription, multiCoins, addr2)},

		{false, NewSubmitProposalMsg(emptyStr, goodDescription, multiCoins, addr1)},      // empty title
		{false, NewSubmitProposalMsg(goodTitle, emptyStr, multiCoins, addr2)},            // empty description
		{false, NewSubmitProposalMsg(emptyStr, emptyStr, someCoins, addr2)},              // invalid title and description
		{false, NewSubmitProposalMsg(emptyStr, emptyStr, emptyCoins, addr1)},             // invalid coins
		{false, NewSubmitProposalMsg(goodTitle, emptyStr, emptyCoins2, addr1)},           // negative coins
		{false, NewSubmitProposalMsg(emptyStr, emptyStr, emptyCoins2, emptyAddr)},        // empty everything
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, minusCoins, emptyAddr)}, // negative coins
		{false, NewSubmitProposalMsg(goodTitle, goodDescription, someEmptyCoins, addr2)}, // empty everything
	}

	for i, msg := range cases {
		err := msg.spMsg.ValidateBasic()
		if msg.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

func TestNewVoteMsg(t *testing.T) {
	yay := "Yes"
	nay := "No"
	abstain := "Abstain"
	other := "Other"
	var emptyStr = ""

	var emptyAddr sdk.Address
	addr1 := sdk.Address([]byte{1, 2})
	addr2 := sdk.Address([]byte{7, 8})

	posPropID := 3
	negPropID := -8

	cases := []struct {
		valid   bool
		voteMsg SubmitProposalMsg
	}{
		{true, NewVoteMsg(posPropID, yay, addr1)},
		{true, NewVoteMsg(posPropID, nay, addr2)},
		{true, NewVoteMsg(posPropID, abstain, addr1)},

		{false, NewVoteMsg(negPropID, yay, addr1)},          // negative ProposalID
		{false, NewVoteMsg(posPropID, other, addr1)},        // other Option
		{false, NewVoteMsg(posPropID, emptyStr, addr1)},     // empty Option
		{false, NewVoteMsg(posPropID, nay, emptyAddr)},      // empty Address
		{false, NewVoteMsg(negPropID, other, emptyAddr)},    // all wrong
		{false, NewVoteMsg(negPropID, emptyStr, emptyAddr)}, // all wrong
	}

	for i, msg := range cases {
		err := msg.voteMsg.ValidateBasic()
		if msg.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}

package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/magiconair/properties/assert"
	crypto "github.com/tendermint/go-crypto"
)

var (
	addr1     = []byte("addr1")
	addr2     = []byte("addr2")
	addr3     = []byte("addr3")
	emptyAddr sdk.Address

	pubkey1     = crypto.GenPrivKeyEd25519().PubKey()
	emptyPubkey crypto.PubKey

	coinPos          = sdk.Coin{"fermion", 1000}
	coinZero         = sdk.Coin{"fermion", 0}
	coinNeg          = sdk.Coin{"fermion", -10000}
	coinPosNotAtoms  = sdk.Coin{"foo", 10000}
	coinZeroNotAtoms = sdk.Coin{"foo", 0}
	coinNegNotAtoms  = sdk.Coin{"foo", -10000}
)

func TestMsgDeclareCandidacy(t *testing.T) {
	tests := []struct {
		name          string
		moniker       string
		identity      string
		website       string
		details       string
		candidateAddr sdk.Address
		pubkey        crypto.PubKey
		bond          sdk.Coin
		expectPass    bool
	}{
		{"basic good", "a", "b", "c", "d", addr1, pubkey1, coinPos, true},
		{"partial description", "", "", "c", "", addr1, pubkey1, coinPos, true},
		{"empty description", "", "", "", "", addr1, pubkey1, coinPos, false},
		{"empty address", "a", "b", "c", "d", emptyAddr, pubkey1, coinPos, false},
		{"empty pubkey", "a", "b", "c", "d", addr1, emptyPubkey, coinPos, true},
		{"empty bond", "a", "b", "c", "d", addr1, pubkey1, coinZero, false},
		{"negative bond", "a", "b", "c", "d", addr1, pubkey1, coinNeg, false},
		{"negative bond", "a", "b", "c", "d", addr1, pubkey1, coinNeg, false},
		{"wrong staking token", "a", "b", "c", "d", addr1, pubkey1, coinPosNotAtoms, false},
	}

	for _, tc := range tests {
		description := Description{
			Moniker:  tc.moniker,
			Identity: tc.identity,
			Website:  tc.website,
			Details:  tc.details,
		}
		msg := NewMsgDeclareCandidacy(tc.candidateAddr, tc.pubkey, tc.bond, description)
		assert.Equal(t, tc.expectPass, msg.ValidateBasic() == nil,
			"test: ", tc.name)
	}
}

func TestMsgEditCandidacy(t *testing.T) {
	tests := []struct {
		name          string
		moniker       string
		identity      string
		website       string
		details       string
		candidateAddr sdk.Address
		expectPass    bool
	}{
		{"basic good", "a", "b", "c", "d", addr1, true},
		{"partial description", "", "", "c", "", addr1, true},
		{"empty description", "", "", "", "", addr1, false},
		{"empty address", "a", "b", "c", "d", emptyAddr, false},
	}

	for _, tc := range tests {
		description := Description{
			Moniker:  tc.moniker,
			Identity: tc.identity,
			Website:  tc.website,
			Details:  tc.details,
		}
		msg := NewMsgEditCandidacy(tc.candidateAddr, description)
		assert.Equal(t, tc.expectPass, msg.ValidateBasic() == nil,
			"test: ", tc.name)
	}
}

func TestMsgDelegate(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.Address
		candidateAddr sdk.Address
		bond          sdk.Coin
		expectPass    bool
	}{
		{"basic good", addr1, addr2, coinPos, true},
		{"self bond", addr1, addr1, coinPos, true},
		{"empty delegator", emptyAddr, addr1, coinPos, false},
		{"empty candidate", addr1, emptyAddr, coinPos, false},
		{"empty bond", addr1, addr2, coinZero, false},
		{"negative bond", addr1, addr2, coinNeg, false},
		{"wrong staking token", addr1, addr2, coinPosNotAtoms, false},
	}

	for _, tc := range tests {
		msg := NewMsgDelegate(tc.delegatorAddr, tc.candidateAddr, tc.bond)
		assert.Equal(t, tc.expectPass, msg.ValidateBasic() == nil,
			"test: ", tc.name)
	}
}

func TestMsgUnbond(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.Address
		candidateAddr sdk.Address
		shares        string
		expectPass    bool
	}{
		{"max unbond", addr1, addr2, "MAX", true},
		{"decimal unbond", addr1, addr2, "0.1", true},
		{"negative decimal unbond", addr1, addr2, "-0.1", false},
		{"zero unbond", addr1, addr2, "0.0", false},
		{"invalid decimal", addr1, addr1, "sunny", false},
		{"empty delegator", emptyAddr, addr1, "0.1", false},
		{"empty candidate", addr1, emptyAddr, "0.1", false},
	}

	for _, tc := range tests {
		msg := NewMsgUnbond(tc.delegatorAddr, tc.candidateAddr, tc.shares)
		assert.Equal(t, tc.expectPass, msg.ValidateBasic() == nil,
			"test: ", tc.name)
	}
}

// TODO introduce with go-amino
//func TestSerializeMsg(t *testing.T) {

//// make sure all types construct properly
//bondAmt := 1234321
//bond := sdk.Coin{Denom: "atom", Amount: int64(bondAmt)}

//tests := []struct {
//tx sdk.Msg
//}{
//{NewMsgDeclareCandidacy(addrs[0], pks[0], bond, Description{})},
//{NewMsgEditCandidacy(addrs[0], Description{})},
//{NewMsgDelegate(addrs[0], addrs[1], bond)},
//{NewMsgUnbond(addrs[0], addrs[1], strconv.Itoa(bondAmt))},
//}

//for i, tc := range tests {
//var tx sdk.Tx
//bs := wire.BinaryBytes(tc.tx)
//err := wire.ReadBinaryBytes(bs, &tx)
//if assert.NoError(t, err, "%d", i) {
//assert.Equal(t, tc.tx, tx, "%d", i)
//}
//}
//}

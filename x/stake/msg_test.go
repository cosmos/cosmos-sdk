package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

var (
	coinPos          = sdk.NewCoin("steak", 1000)
	coinZero         = sdk.NewCoin("steak", 0)
	coinNeg          = sdk.NewCoin("steak", -10000)
	coinPosNotAtoms  = sdk.NewCoin("foo", 10000)
	coinZeroNotAtoms = sdk.NewCoin("foo", 0)
	coinNegNotAtoms  = sdk.NewCoin("foo", -10000)
)

// test ValidateBasic for MsgCreateValidator
func TestMsgCreateValidator(t *testing.T) {
	tests := []struct {
		name, moniker, identity, website, details string
		validatorAddr                             sdk.Address
		pubkey                                    crypto.PubKey
		bond                                      sdk.Coin
		expectPass                                bool
	}{
		{"basic good", "a", "b", "c", "d", addrs[0], pks[0], coinPos, true},
		{"partial description", "", "", "c", "", addrs[0], pks[0], coinPos, true},
		{"empty description", "", "", "", "", addrs[0], pks[0], coinPos, false},
		{"empty address", "a", "b", "c", "d", emptyAddr, pks[0], coinPos, false},
		{"empty pubkey", "a", "b", "c", "d", addrs[0], emptyPubkey, coinPos, true},
		{"empty bond", "a", "b", "c", "d", addrs[0], pks[0], coinZero, false},
		{"negative bond", "a", "b", "c", "d", addrs[0], pks[0], coinNeg, false},
		{"negative bond", "a", "b", "c", "d", addrs[0], pks[0], coinNeg, false},
		{"wrong staking token", "a", "b", "c", "d", addrs[0], pks[0], coinPosNotAtoms, false},
	}

	for _, tc := range tests {
		description := NewDescription(tc.moniker, tc.identity, tc.website, tc.details)
		msg := NewMsgCreateValidator(tc.validatorAddr, tc.pubkey, tc.bond, description)
		if tc.expectPass {
			assert.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			assert.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgEditValidator
func TestMsgEditValidator(t *testing.T) {
	tests := []struct {
		name, moniker, identity, website, details string
		validatorAddr                             sdk.Address
		expectPass                                bool
	}{
		{"basic good", "a", "b", "c", "d", addrs[0], true},
		{"partial description", "", "", "c", "", addrs[0], true},
		{"empty description", "", "", "", "", addrs[0], false},
		{"empty address", "a", "b", "c", "d", emptyAddr, false},
	}

	for _, tc := range tests {
		description := NewDescription(tc.moniker, tc.identity, tc.website, tc.details)
		msg := NewMsgEditValidator(tc.validatorAddr, description)
		if tc.expectPass {
			assert.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			assert.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgDelegate
func TestMsgDelegate(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.Address
		validatorAddr sdk.Address
		bond          sdk.Coin
		expectPass    bool
	}{
		{"basic good", addrs[0], addrs[1], coinPos, true},
		{"self bond", addrs[0], addrs[0], coinPos, true},
		{"empty delegator", emptyAddr, addrs[0], coinPos, false},
		{"empty validator", addrs[0], emptyAddr, coinPos, false},
		{"empty bond", addrs[0], addrs[1], coinZero, false},
		{"negative bond", addrs[0], addrs[1], coinNeg, false},
		{"wrong staking token", addrs[0], addrs[1], coinPosNotAtoms, false},
	}

	for _, tc := range tests {
		msg := NewMsgDelegate(tc.delegatorAddr, tc.validatorAddr, tc.bond)
		if tc.expectPass {
			assert.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			assert.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgUnbond(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.Address
		validatorAddr sdk.Address
		shares        string
		expectPass    bool
	}{
		{"max unbond", addrs[0], addrs[1], "MAX", true},
		{"decimal unbond", addrs[0], addrs[1], "0.1", true},
		{"negative decimal unbond", addrs[0], addrs[1], "-0.1", false},
		{"zero unbond", addrs[0], addrs[1], "0.0", false},
		{"invalid decimal", addrs[0], addrs[0], "sunny", false},
		{"empty delegator", emptyAddr, addrs[0], "0.1", false},
		{"empty validator", addrs[0], emptyAddr, "0.1", false},
	}

	for _, tc := range tests {
		msg := NewMsgUnbond(tc.delegatorAddr, tc.validatorAddr, tc.shares)
		if tc.expectPass {
			assert.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			assert.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
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
//{NewMsgCreateValidator(addrs[0], pks[0], bond, Description{})},
//{NewMsgEditValidator(addrs[0], Description{})},
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

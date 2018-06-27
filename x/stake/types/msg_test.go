package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

var (
	coinPos  = sdk.Coin{"steak", sdk.NewInt(1000)}
	coinZero = sdk.Coin{"steak", sdk.NewInt(0)}
	coinNeg  = sdk.Coin{"steak", sdk.NewInt(-10000)}
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
		{"basic good", "a", "b", "c", "d", addr1, pk1, coinPos, true},
		{"partial description", "", "", "c", "", addr1, pk1, coinPos, true},
		{"empty description", "", "", "", "", addr1, pk1, coinPos, false},
		{"empty address", "a", "b", "c", "d", emptyAddr, pk1, coinPos, false},
		{"empty pubkey", "a", "b", "c", "d", addr1, emptyPubkey, coinPos, true},
		{"empty bond", "a", "b", "c", "d", addr1, pk1, coinZero, false},
		{"negative bond", "a", "b", "c", "d", addr1, pk1, coinNeg, false},
		{"negative bond", "a", "b", "c", "d", addr1, pk1, coinNeg, false},
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
		{"basic good", "a", "b", "c", "d", addr1, true},
		{"partial description", "", "", "c", "", addr1, true},
		{"empty description", "", "", "", "", addr1, false},
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
		{"basic good", addr1, addr2, coinPos, true},
		{"self bond", addr1, addr1, coinPos, true},
		{"empty delegator", emptyAddr, addr1, coinPos, false},
		{"empty validator", addr1, emptyAddr, coinPos, false},
		{"empty bond", addr1, addr2, coinZero, false},
		{"negative bond", addr1, addr2, coinNeg, false},
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
func TestMsgBeginRedelegate(t *testing.T) {
	tests := []struct {
		name             string
		delegatorAddr    sdk.Address
		validatorSrcAddr sdk.Address
		validatorDstAddr sdk.Address
		sharesAmount     sdk.Rat
		expectPass       bool
	}{
		{"regular", addr1, addr2, addr3, sdk.NewRat(1, 10), true},
		{"negative decimal", addr1, addr2, addr3, sdk.NewRat(-1, 10), false},
		{"zero amount", addr1, addr2, addr3, sdk.ZeroRat(), false},
		{"empty delegator", emptyAddr, addr1, addr3, sdk.NewRat(1, 10), false},
		{"empty source validator", addr1, emptyAddr, addr3, sdk.NewRat(1, 10), false},
		{"empty destination validator", addr1, addr2, emptyAddr, sdk.NewRat(1, 10), false},
	}

	for _, tc := range tests {
		msg := NewMsgBeginRedelegate(tc.delegatorAddr, tc.validatorSrcAddr, tc.validatorDstAddr, tc.sharesAmount)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgCompleteRedelegate(t *testing.T) {
	tests := []struct {
		name             string
		delegatorAddr    sdk.Address
		validatorSrcAddr sdk.Address
		validatorDstAddr sdk.Address
		expectPass       bool
	}{
		{"regular", addr1, addr2, addr3, true},
		{"empty delegator", emptyAddr, addr1, addr3, false},
		{"empty source validator", addr1, emptyAddr, addr3, false},
		{"empty destination validator", addr1, addr2, emptyAddr, false},
	}

	for _, tc := range tests {
		msg := NewMsgCompleteRedelegate(tc.delegatorAddr, tc.validatorSrcAddr, tc.validatorDstAddr)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgBeginUnbonding(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.Address
		validatorAddr sdk.Address
		sharesAmount  sdk.Rat
		expectPass    bool
	}{
		{"regular", addr1, addr2, sdk.NewRat(1, 10), true},
		{"negative decimal", addr1, addr2, sdk.NewRat(-1, 10), false},
		{"zero amount", addr1, addr2, sdk.ZeroRat(), false},
		{"empty delegator", emptyAddr, addr1, sdk.NewRat(1, 10), false},
		{"empty validator", addr1, emptyAddr, sdk.NewRat(1, 10), false},
	}

	for _, tc := range tests {
		msg := NewMsgBeginUnbonding(tc.delegatorAddr, tc.validatorAddr, tc.sharesAmount)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgCompleteUnbonding(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.Address
		validatorAddr sdk.Address
		expectPass    bool
	}{
		{"regular", addr1, addr2, true},
		{"empty delegator", emptyAddr, addr1, false},
		{"empty validator", addr1, emptyAddr, false},
	}

	for _, tc := range tests {
		msg := NewMsgCompleteUnbonding(tc.delegatorAddr, tc.validatorAddr)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
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
//{NewMsgCreateValidator(addr1, pk1, bond, Description{})},
//{NewMsgEditValidator(addr1, Description{})},
//{NewMsgDelegate(addr1, addr2, bond)},
//{NewMsgUnbond(addr1, addr2, strconv.Itoa(bondAmt))},
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

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	// tmtypes "github.com/tendermint/tendermint/types"

	// sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)


func TestMsgProposeCreateValidator(t *testing.T) {
	tests := []struct{
	title, description string
	newValidator MsgCreateValidator
	expectPass                                                 bool
}{
	{"foo", "bar",MsgCreateValidator{stakingtypes.Description{"basic good", "a", "b", "c", "d"}, valAddr1, pk1}, true},
	{"bar","foo",MsgCreateValidator{stakingtypes.Description{"partial description", "", "", "c", ""}, valAddr1, pk1}, true},
	{"","Alcie",MsgCreateValidator{stakingtypes.Description{"empty description", "", "", "", ""}, valAddr1, pk1}, false},
	{"bob","",MsgCreateValidator{stakingtypes.Description{"empty address", "a", "b", "c", "d"}, emptyAddr, pk1}, false},
	{"","simon",MsgCreateValidator{stakingtypes.Description{"empty pubkey", "a", "b", "c", "d"}, valAddr1, emptyPubkey}, false},
}
for _, tc := range tests {
	msg := NewMsgProposeCreateValidator(tc.title, tc.description,tc.newValidator.ValidatorAddress, tc.newValidator.PubKey, tc.newValidator.Description)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.title)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.title)
		}
	}
}

// func TestMsg
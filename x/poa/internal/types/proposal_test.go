package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	// tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMsgProposeCreateValidator(t *testing.T) {
	title, description := "foo", "bar"
	createVal := MsgCreateValidator{stakingtypes.Description{"basic good", "a", "b", "c", "d"}, valAddr1, pk1}
	newVal := NewMsgProposeCreateValidator(title, description, createVal.ValidatorAddress,
		createVal.PubKey, createVal.Description)
	require.Equal(t, title, newVal.GetTitle())
	require.Equal(t, description, newVal.GetDescription())
	require.Equal(t, RouterKey, newVal.ProposalRoute())
	require.Equal(t, ProposeCreateValidator, newVal.ProposalType())
}

func TestMsgProposeCreateValidatorValidation(t *testing.T) {
	//nolint: govet
	tests := []struct {
		title, description string
		newValidator       MsgCreateValidator
		expectPass         bool
	}{
		{"foo", "bar", MsgCreateValidator{stakingtypes.Description{"basic good", "a", "b", "c", "d"}, valAddr1, pk1}, true},
		{"bar", "foo", MsgCreateValidator{stakingtypes.Description{"partial description", "", "", "c", ""}, valAddr1, pk1}, true},
		{"", "Alcie", MsgCreateValidator{stakingtypes.Description{"empty description", "", "", "", ""}, valAddr1, pk1}, false},
		{"bob", "", MsgCreateValidator{stakingtypes.Description{"empty address", "a", "b", "c", "d"}, emptyAddr, pk1}, false},
		{"", "simon", MsgCreateValidator{stakingtypes.Description{"empty pubkey", "a", "b", "c", "d"}, valAddr1, emptyPubkey}, false},
	}
	for _, tc := range tests {
		msg := NewMsgProposeCreateValidator(tc.title, tc.description, tc.newValidator.ValidatorAddress, tc.newValidator.PubKey, tc.newValidator.Description)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.title)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.title)
		}
	}
}
func TestMsgProposeIncreaseWeight(t *testing.T) {
	title, description := "foo", "bar"
	createdVal := ValidatorIncreaseWeight{valAddr1, sdk.NewInt(10)}
	updatedVal := NewMsgProposeIncreaseWeight(title, description, createdVal)

	require.Equal(t, title, updatedVal.GetTitle())
	require.Equal(t, description, updatedVal.GetDescription())
	require.Equal(t, RouterKey, updatedVal.ProposalRoute())
	require.Equal(t, ProposeIncreaseWeight, updatedVal.ProposalType())
}

func TestMsgProposeIncreaseweightValidation(t *testing.T) {
	tests := []struct {
		title, description string
		validatorIncrease  ValidatorIncreaseWeight
		expectPass         bool
	}{
		{"foo", "bar", ValidatorIncreaseWeight{valAddr1, sdk.NewInt(5)}, true},
		{"", "bar", ValidatorIncreaseWeight{valAddr1, sdk.NewInt(5)}, false},
		{"foo", "", ValidatorIncreaseWeight{valAddr1, sdk.NewInt(5)}, false},
	}
	for _, tc := range tests {
		msg := NewMsgProposeIncreaseWeight(tc.title, tc.description, tc.validatorIncrease)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.title)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.title)
		}
	}
}

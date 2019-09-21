package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMsgCreateValidator(t *testing.T) {
	desc := stakingtypes.Description{"basic good", "a", "b", "c", "d"}
	var msg = NewMsgCreateValidator(valAddr1, pk1, desc)

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "create_validator")
}

// test ValidateBasic for MsgCreateValidator
func TestMsgCreateValidatorValidation(t *testing.T) {

	tests := []struct {
		name, moniker, identity, website, securityContact, details string
		validatorAddr                                              sdk.ValAddress
		pubkey                                                     crypto.PubKey
		expectPass                                                 bool
	}{
		{"basic good", "a", "b", "c", "d", "e", valAddr1, pk1, true},
		{"partial description", "", "", "c", "", "", valAddr1, pk1, true},
		{"empty description", "", "", "", "", "", valAddr1, pk1, false},
		{"empty address", "a", "b", "c", "d", "e", emptyAddr, pk1, false},
		{"empty pubkey", "a", "b", "c", "d", "e", valAddr1, emptyPubkey, true},
	}

	for _, tc := range tests {
		description := stakingtypes.NewDescription(tc.moniker, tc.identity, tc.website, tc.securityContact, tc.details)
		msg := NewMsgCreateValidator(tc.validatorAddr, tc.pubkey, description)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

func TestMsgEditValidator(t *testing.T) {
	desc := stakingtypes.Description{"basic good", "a", "b", "c", "d"}
	var msg = NewMsgEditValidator(valAddr1, desc)

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "edit_validator")
}

// test ValidateBasic for MsgEditValidator
func TestMsgEditValidatorValidation(t *testing.T) {
	tests := []struct {
		name, moniker, identity, website, securityContact, details string
		validatorAddr                                              sdk.ValAddress
		expectPass                                                 bool
	}{
		{"basic good", "a", "b", "c", "d", "e", valAddr1, true},
		{"partial description", "", "", "c", "", "", valAddr1, true},
		{"empty description", "", "", "", "", "", valAddr1, false},
		{"empty address", "a", "b", "c", "d", "e", emptyAddr, false},
	}

	for _, tc := range tests {
		description := stakingtypes.NewDescription(tc.moniker, tc.identity, tc.website, tc.securityContact, tc.details)

		msg := NewMsgEditValidator(tc.validatorAddr, description)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

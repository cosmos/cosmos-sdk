package types

import (
	"testing"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	// yaml "gopkg.in/yaml.v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestValidatorTestEquivalent(t *testing.T) {
	val1 := NewValidator(valAddr1, pk1, stakingtypes.Description{})
	val2 := NewValidator(valAddr1, pk1, stakingtypes.Description{})

	ok := val1.TestEquivalent(val2)
	require.True(t, ok)

	val2 = NewValidator(valAddr2, pk2, stakingtypes.Description{})

	ok = val1.TestEquivalent(val2)
	require.False(t, ok)
}

func TestUpdateDescription(t *testing.T) {
	d1 := stakingtypes.Description{
		Website: "https://validator.cosmos",
		Details: "Test validator",
	}

	d2 := stakingtypes.Description{
		Moniker:  stakingtypes.DoNotModifyDesc,
		Identity: stakingtypes.DoNotModifyDesc,
		Website:  stakingtypes.DoNotModifyDesc,
		Details:  stakingtypes.DoNotModifyDesc,
	}

	d3 := stakingtypes.Description{
		Moniker:  "",
		Identity: "",
		Website:  "",
		Details:  "",
	}

	d, err := d1.UpdateDescription(d2)
	require.Nil(t, err)
	require.Equal(t, d, d1)

	d, err = d1.UpdateDescription(d3)
	require.Nil(t, err)
	require.Equal(t, d, d3)
}

func TestABCIValidatorUpdate(t *testing.T) {
	validator := NewValidator(valAddr1, pk1, stakingtypes.Description{})

	abciVal := validator.ABCIValidatorUpdate()
	require.Equal(t, tmtypes.TM2PB.PubKey(validator.ConsPubKey), abciVal.PubKey)
	require.Equal(t, validator.GetBondedWeight().Int64(), abciVal.Power)
}

func TestABCIValidatorUpdateZero(t *testing.T) {
	validator := NewValidator(valAddr1, pk1, stakingtypes.Description{})

	abciVal := validator.ABCIValidatorUpdateZero()
	require.Equal(t, tmtypes.TM2PB.PubKey(validator.ConsPubKey), abciVal.PubKey)
	require.Equal(t, int64(0), abciVal.Power)
}

func TestRemoveTokens(t *testing.T) {
	valPubKey := pk1
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

	validator := Validator{
		OperatorAddress: valAddr,
		ConsPubKey:      valPubKey,
		Status:          sdk.Bonded,
		Weight:          sdk.NewInt(10),
	}

	// remove tokens and test check everything
	validator = validator.RemoveWeight(sdk.NewInt(10))
	require.Equal(t, int64(0), validator.Weight.Int64())

	// update validator to from bonded -> unbonded
	validator = validator.UpdateStatus(sdk.Unbonded)
	require.Equal(t, sdk.Unbonded, validator.Status)

	// validator = validator.RemoveWeight(sdk.NewInt(1))
	// require.Panics(t, func() { validator.RemoveWeight(sdk.NewInt(1)) })
}

func TestUpdateStatus(t *testing.T) {
	validator := NewValidator(sdk.ValAddress(pk1.Address().Bytes()), pk1, stakingtypes.Description{})
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(10), validator.Weight.Int64())

	// Unbonded to Bonded
	validator = validator.UpdateStatus(sdk.Bonded)
	require.Equal(t, sdk.Bonded, validator.Status)

	// Bonded to Unbonding
	validator = validator.UpdateStatus(sdk.Unbonding)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// Unbonding to Bonded
	validator = validator.UpdateStatus(sdk.Bonded)
	require.Equal(t, sdk.Bonded, validator.Status)
}

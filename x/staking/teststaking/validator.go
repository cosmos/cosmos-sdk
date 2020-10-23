package teststaking

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewValidator is a testing helper method to create validators in tests
func NewValidator(t *testing.T, operator sdk.ValAddress, pubKey crypto.PubKey) types.Validator {
	v, err := types.NewValidator(operator, pubKey, types.Description{})
	require.NoError(t, err)
	return v
}

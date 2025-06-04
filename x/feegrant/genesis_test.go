package feegrant

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDuplicateGrantsInGenesis(t *testing.T) {
	// Create dummy addresses for test
	granter := sdk.AccAddress("granter_address____").String()
	grantee := sdk.AccAddress("grantee_address____").String()

	// Create a BasicAllowance for testing
	allowance := &BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(100))),
	}

	any, err := codectypes.NewAnyWithValue(allowance)
	assert.NilError(t, err)

	// Create Genesis state with duplicate allowances
	genesisState := &GenesisState{
		Allowances: []Grant{
			{
				Granter:   granter,
				Grantee:   grantee,
				Allowance: any,
			},
			{
				Granter:   granter,
				Grantee:   grantee,
				Allowance: any,
			},
		},
	}

	// Validation should fail with duplicate feegrant error
	err = ValidateGenesis(*genesisState)
	assert.ErrorContains(t, err, "duplicate feegrant found")
}

package v4_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type mockGovParamSubspace struct {
	depositParams v1.DepositParams
}

var _ types.ParamSubspace = (*mockGovParamSubspace)(nil)

func (mp mockGovParamSubspace) Get(ctx sdk.Context, key []byte, param interface{}) {
	depositParams, ok := (param).(*v1.DepositParams)
	if !ok {
		panic(fmt.Sprintf("could not cast %v to deposit params", param))
	}
	*depositParams = mp.depositParams
}
func (mp *mockGovParamSubspace) Set(ctx sdk.Context, key []byte, param interface{}) {
	depositParams, ok := (param).(*v1.DepositParams)
	if !ok {
		panic(fmt.Sprintf("could not cast %v to deposit params", param))
	}
	mp.depositParams = *depositParams
}

func TestGovStoreMigrationToV4ConsensusVersion(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	paramSubspace := mockGovParamSubspace{}

	require.NoError(t, v4.MigrateStore(ctx, &paramSubspace))

	// Make sure the new param is set.
	var depositParams v1.DepositParams
	paramSubspace.Get(ctx, v1.ParamStoreKeyDepositParams, &depositParams)
	require.Equal(t, v4.MinInitialDepositRatio, depositParams.MinInitialDepositRatio)
}

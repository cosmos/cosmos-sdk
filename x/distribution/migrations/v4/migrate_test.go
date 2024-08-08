package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/x/distribution"
	v4 "cosmossdk.io/x/distribution/migrations/v4"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestMigration(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, distribution.AppModule{}).Codec
	ctx := coretesting.Context()
	storeService := coretesting.KVStoreService(ctx, "distribution")

	env := runtime.NewEnvironment(storeService, coretesting.NewNopLogger())

	addr1 := secp256k1.GenPrivKey().PubKey().Address()
	consAddr1 := sdk.ConsAddress(addr1)

	// Set and check the previous proposer
	err := v4.SetPreviousProposerConsAddr(ctx, storeService, cdc, consAddr1)
	require.NoError(t, err)

	gotAddr, err := v4.GetPreviousProposerConsAddr(ctx, storeService, cdc)
	require.NoError(t, err)
	require.Equal(t, consAddr1, gotAddr)

	err = v4.MigrateStore(ctx, env, cdc)
	require.NoError(t, err)

	// Check that the previous proposer has been removed
	_, err = v4.GetPreviousProposerConsAddr(ctx, storeService, cdc)
	require.ErrorContains(t, err, "previous proposer not set")
}

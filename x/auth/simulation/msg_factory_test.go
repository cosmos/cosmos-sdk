package simulation_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/simulation"
)

type moduleAccountSource struct{}

func (moduleAccountSource) GetModuleAddress(moduleName string) sdk.AccAddress {
	return address.Module(moduleName)
}

func TestMsgUpdateParamsFactory(t *testing.T) {
	codec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	accs := simtypes.RandomAccounts(rand.New(rand.NewSource(1)), 3)
	factory := simulation.MsgUpdateParamsFactory()

	seenEnabled := map[bool]bool{}
	for seed := int64(0); seed < 20; seed++ {
		r := rand.New(rand.NewSource(seed))
		reporter := simsx.NewBasicSimulationReporter()
		testData := simsx.NewChainDataSource(context.Background(), r, moduleAccountSource{}, nil, codec, accs...)

		signers, msg := factory(context.Background(), testData, reporter)
		require.False(t, reporter.IsSkipped())
		require.Empty(t, signers)
		require.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msg.Authority)

		params := msg.Params
		require.NoError(t, params.Validate())
		require.GreaterOrEqual(t, params.PubKeyChangeCost, uint64(1000))
		require.LessOrEqual(t, params.PubKeyChangeCost, uint64(40000))
		seenEnabled[params.PubKeyChangeEnabled] = true
	}
	// Param proposals both enable and disable pubkey changes.
	require.True(t, seenEnabled[true])
	require.True(t, seenEnabled[false])
}

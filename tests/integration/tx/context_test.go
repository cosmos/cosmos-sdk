package tx

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	_ "cosmossdk.io/x/accounts"
	"cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/tx/internal"
	"github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/pulsar/testpb"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func ProvideCustomGetSigner() signing.CustomGetSigner {
	return internal.TestRepeatedFieldsSigner
}

func TestDefineCustomGetSigners(t *testing.T) {
	var interfaceRegistry codectypes.InterfaceRegistry
	_, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AccountsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.BankModule(),
				configurator.ConsensusModule(),
			),
			depinject.Supply(log.NewNopLogger()),
			depinject.Provide(ProvideCustomGetSigner),
		),
		&interfaceRegistry,
	)
	require.NoError(t, err)
	require.NotNil(t, interfaceRegistry)

	msg := &testpb.TestRepeatedFields{
		NullableDontOmitempty: []*testpb.Streng{
			{Value: "foo"},
			{Value: "bar"},
		},
	}
	signers, err := interfaceRegistry.SigningContext().GetSigners(msg)
	require.NoError(t, err)
	require.Equal(t, [][]byte{[]byte("bar")}, signers)

	// Reset and provider no CustomGetSigners. Consequently, validation will fail and depinject will return an error
	_, err = simtestutil.SetupAtGenesis(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.BankModule(),
				configurator.ConsensusModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&interfaceRegistry,
	)
	require.ErrorContains(t, err, "use DefineCustomGetSigners")
}

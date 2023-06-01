package tx

import (
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/x/tx/signing"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func SetCustomGetSigners(registry codectypes.InterfaceRegistry) {
	signing.DefineCustomGetSigners(registry.SigningContext(), func(msg *bankv1beta1.SendAuthorization) ([][]byte, error) {
		// arbitrary logic
		signer := msg.AllowList[1]
		return [][]byte{[]byte(signer)}, nil
	})
}

func TestDefineCustomGetSigners(t *testing.T) {
	var interfaceRegistry codectypes.InterfaceRegistry
	_, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.ParamsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.BankModule(),
				configurator.ConsensusModule(),
			),
			depinject.Supply(log.NewNopLogger()),
			depinject.Invoke(SetCustomGetSigners),
		),
		&interfaceRegistry,
	)
	require.NoError(t, err)
	require.NotNil(t, interfaceRegistry)

	sendAuth := &bankv1beta1.SendAuthorization{
		AllowList: []string{"foo", "bar"},
	}
	signers, err := interfaceRegistry.SigningContext().GetSigners(sendAuth)
	require.Equal(t, [][]byte{[]byte("bar")}, signers)

	// reset without invoker, no custom signer.
	_, err = simtestutil.SetupAtGenesis(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.ParamsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.BankModule(),
				configurator.ConsensusModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&interfaceRegistry,
	)
	require.NoError(t, err)
	require.NotNil(t, interfaceRegistry)

	_, err = interfaceRegistry.SigningContext().GetSigners(sendAuth)
	require.ErrorContains(t, err, "use DefineCustomGetSigners")
}

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

func ProvideCustomGetSigners() signing.CustomGetSignersImpl[*bankv1beta1.SendAuthorization] {
	return func(msg *bankv1beta1.SendAuthorization) ([][]byte, error) {
		return [][]byte{[]byte("foo")}, nil
	}
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
				configurator.GovModule(),
				configurator.DistributionModule(),
				configurator.ConsensusModule(),
			),
			depinject.Supply(log.NewNopLogger()),
			depinject.Provide(ProvideCustomGetSigners),
		),
		&interfaceRegistry,
	)
	require.NoError(t, err)
	require.NotNil(t, interfaceRegistry)
}

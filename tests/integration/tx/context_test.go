package tx

import (
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/x/tx/signing"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func ProvideCustomGetSigners() signing.CustomGetSigner {
	return signing.CustomGetSigner{
		MsgType: proto.MessageName(&bankv1beta1.SendAuthorization{}),
		Fn: func(msg proto.Message) ([][]byte, error) {
			sendAuth := msg.(*bankv1beta1.SendAuthorization)
			// arbitrary logic
			signer := sendAuth.AllowList[1]
			return [][]byte{[]byte(signer)}, nil
		},
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
				configurator.ConsensusModule(),
			),
			depinject.Supply(log.NewNopLogger()),
			depinject.Provide(ProvideCustomGetSigners),
		),
		&interfaceRegistry,
	)
	require.NoError(t, err)
	require.NotNil(t, interfaceRegistry)

	sendAuth := &bankv1beta1.SendAuthorization{
		AllowList: []string{"foo", "bar"},
	}
	signers, err := interfaceRegistry.SigningContext().GetSigners(sendAuth)
	require.NoError(t, err)
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

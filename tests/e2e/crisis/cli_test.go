//go:build e2e
// +build e2e

package crisis

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/client/cli"
	"github.com/cosmos/gogoproto/proto"
	"gotest.tools/v3/assert"
)

type fixture struct {
	cfg     network.Config
	network *network.Network
}

func initFixture(t *testing.T) *fixture {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1

	network, err := network.New(t, t.TempDir(), cfg)
	assert.NilError(t, err)
	assert.NilError(t, network.WaitForNextBlock())

	return &fixture{
		cfg:     cfg,
		network: network,
	}
}

func TestNewMsgVerifyInvariantTxCmd(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.network.Cleanup()

	val := f.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"missing module",
			[]string{
				"", "total-supply",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid module name", 0, nil,
		},
		{
			"missing invariant route",
			[]string{
				"bank", "",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdkmath.NewInt(10))).String()),
			},
			"invalid invariant route", 0, nil,
		},
		{
			"valid traclientCtx.Codec.UnmarshalJSON(out.Bytes(), nsaction",
			[]string{
				"bank", "total-supply",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdkmath.NewInt(10))).String()),
			},
			"", 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.NewMsgVerifyInvariantTxCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				assert.ErrorContains(t, err, tc.expectErrMsg)
			} else {
				assert.NilError(t, err)
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				assert.NilError(t, clitestutil.CheckTxCode(f.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

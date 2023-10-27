package testutil

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/client/cli"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	if testing.Short() {
		s.T().Skip("skipping test in unit-tests mode.")
	}

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

// TestBlockResults tests that the validator updates correctly show when
// calling the /block_results RPC endpoint.
// ref: https://github.com/cosmos/cosmos-sdk/issues/7401.
func (s *E2ETestSuite) TestBlockResults() {
	require := s.Require()
	val := s.network.Validators[0]

	// Create new account in the keyring.
	k, _, err := val.ClientCtx.Keyring.NewMnemonic("NewDelegator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)
	pub, err := k.GetPubKey()
	require.NoError(err)
	newAddr := sdk.AccAddress(pub.Address())

	msgSend := &banktypes.MsgSend{
		FromAddress: val.Address.String(),
		ToAddress:   newAddr.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(200))),
	}

	// Send some funds to the new account.
	_, err = clitestutil.SubmitTestTx(
		val.ClientCtx,
		msgSend,
		val.Address,
		clitestutil.TestTxConfig{},
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	msgDel := &stakingtypes.MsgDelegate{
		DelegatorAddress: newAddr.String(),
		ValidatorAddress: val.ValAddress.String(),
		Amount:           sdk.NewCoin(s.cfg.BondDenom, math.NewInt(150)),
	}

	// create a delegation from the new account to validator `val`.
	_, err = clitestutil.SubmitTestTx(
		val.ClientCtx,
		msgDel,
		newAddr,
		clitestutil.TestTxConfig{},
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	// Create a HTTP rpc client.
	rpcClient, err := http.New(val.RPCAddress, "/websocket")
	require.NoError(err)

	// Loop until we find a block result with the correct validator updates.
	// By experience, it happens around 2 blocks after `delHeight`.
	_ = s.network.RetryForBlocks(func() error {
		latestHeight, err := s.network.LatestHeight()
		require.NoError(err)
		res, err := rpcClient.BlockResults(context.Background(), &latestHeight)
		if err != nil {
			return err
		}

		if len(res.ValidatorUpdates) == 0 {
			return errors.New("validator update not found yet")
		}

		valUpdate := res.ValidatorUpdates[0]
		require.Equal(
			valUpdate.GetPubKey().Sum.(*crypto.PublicKey_Ed25519).Ed25519,
			val.PubKey.Bytes(),
		)

		return nil
	}, 10)
}

func (s *E2ETestSuite) TestNewRotateConsKeyCmd() {
	val := s.network.Validators[0]

	cmd := cli.NewRotateConsKeyCmd()
	validPubKey := `{"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="}`

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong pubkey",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=%s", cli.FlagPubKey, "wrongPubKey"),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true,
		},
		{
			"missing pubkey",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true,
		},
		{
			"valid case",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=%s", cli.FlagPubKey, validPubKey),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, tc.args)
			fmt.Printf("out: %v\n", out)
			fmt.Printf("err: %v\n", err)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				resp := &sdk.TxResponse{}
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), resp)
				s.Require().NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

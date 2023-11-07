package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network network.NetworkI
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
	val := s.network.GetValidators()[0]

	// Create new account in the keyring.
	k, _, err := val.GetClientCtx().Keyring.NewMnemonic("NewDelegator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)
	pub, err := k.GetPubKey()
	require.NoError(err)
	newAddr := sdk.AccAddress(pub.Address())

	msgSend := &banktypes.MsgSend{
		FromAddress: val.GetAddress().String(),
		ToAddress:   newAddr.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(200))),
	}

	// Send some funds to the new account.
	_, err = clitestutil.SubmitTestTx(
		val.GetClientCtx(),
		msgSend,
		val.GetAddress(),
		clitestutil.TestTxConfig{},
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	msgDel := &stakingtypes.MsgDelegate{
		DelegatorAddress: newAddr.String(),
		ValidatorAddress: val.GetValAddress().String(),
		Amount:           sdk.NewCoin(s.cfg.BondDenom, math.NewInt(150)),
	}

	// create a delegation from the new account to validator `val`.
	_, err = clitestutil.SubmitTestTx(
		val.GetClientCtx(),
		msgDel,
		newAddr,
		clitestutil.TestTxConfig{},
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	// Create a HTTP rpc client.
	rpcClient, err := http.New(val.GetRPCAddress(), "/websocket")
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
			val.GetPubKey().Bytes(),
		)

		return nil
	}, 10)
}

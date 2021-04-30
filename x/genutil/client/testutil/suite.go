package testutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.network = network.NewForTesting(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGenTxCmd() {
	val := s.network.Validators[0]
	dir := s.T().TempDir()

	cmd := cli.GenTxCmd(
		simapp.ModuleBasics,
		val.ClientCtx.TxConfig, banktypes.GenesisBalancesIterator{}, val.ClientCtx.HomeDir)

	_, out := testutil.ApplyMockIO(cmd)
	clientCtx := val.ClientCtx.WithOutput(out)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	amount := sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))
	genTxFile := filepath.Join(dir, "myTx")
	cmd.SetArgs([]string{
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.network.Config.ChainID),
		fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, genTxFile),
		val.Moniker,
		amount.String(),
	})

	err := cmd.ExecuteContext(ctx)
	s.Require().NoError(err)

	// validate generated transaction.
	open, err := os.Open(genTxFile)
	s.Require().NoError(err)

	all, err := ioutil.ReadAll(open)
	s.Require().NoError(err)

	tx, err := val.ClientCtx.TxConfig.TxJSONDecoder()(all)
	s.Require().NoError(err)

	msgs := tx.GetMsgs()
	s.Require().Len(msgs, 1)

	s.Require().Equal(types.TypeMsgCreateValidator, msgs[0].(legacytx.LegacyMsg).Type())
	s.Require().Equal([]sdk.AccAddress{val.Address}, msgs[0].GetSigners())
	s.Require().Equal(amount, msgs[0].(*types.MsgCreateValidator).Value)
	err = tx.ValidateBasic()
	s.Require().NoError(err)
}

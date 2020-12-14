// +build norace ledger,test_ledger_mock

package cli_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
)

type LedgerIntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *LedgerIntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *LedgerIntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *LedgerIntegrationTestSuite) TestSendWithLedger() {
	val0 := s.network.Validators[0]

	cmd := keys.AddKeyCommand()
	cmd.Flags().AddFlagSet(keys.Commands("home").PersistentFlags())

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	kbHome := s.T().TempDir()

	clientCtx := val0.ClientCtx.WithKeyringDir(kbHome)

	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	cmd.SetArgs([]string{
		"keyname1",
		fmt.Sprintf("--%s=true", flags.FlagUseLedger),
		fmt.Sprintf("--%s=%s", cli.OutputFlag, keys.OutputFormatText),
		fmt.Sprintf("--%s=%s", flags.FlagKeyAlgorithm, string(hd.Secp256k1Type)),
		fmt.Sprintf("--%s=%d", "coin-type", sdk.CoinType),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	mockIn.Reset("test1234\ntest1234\n")

	s.Require().NoError(cmd.ExecuteContext(ctx))

	// Now check that it has been stored properly
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn)
	s.Require().NoError(err)
	s.Require().NotNil(kb)
	s.T().Cleanup(func() {
		_ = kb.Delete("keyname1")
	})

	mockIn.Reset("test1234\n")
	key1, err := kb.Key("keyname1")
	s.Require().NoError(err)
	s.Require().NotNil(key1)

	s.Require().Equal("keyname1", key1.GetName())
	s.Require().Equal(keyring.TypeLedger, key1.GetType())

	kb, err = keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn)
	s.Require().NoError(err)
	clientCtx = clientCtx.WithKeyring(kb)
	_, err = clientCtx.Keyring.Key("keyname1")
	s.Require().NoError(err)

	// Send some funds to keyname1.
	_, err = banktestutil.MsgSendExec(val0.ClientCtx, val0.Address, key1.GetAddress(), sdk.NewCoin(fmt.Sprintf("%stoken", val0.Moniker), sdk.NewInt(10)))
	s.Require().NoError(err)

	sendCmd := bankcli.NewSendTxCmd()
	mockIn = testutil.ApplyMockIODiscardOutErr(sendCmd)
	_, err = clitestutil.ExecTestCLICmd(
		clientCtx, sendCmd,
		[]string{
			key1.GetName(), key1.GetAddress().String(), sdk.NewCoin(fmt.Sprintf("%stoken", val0.Moniker), sdk.NewInt(1)).String(),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		},
	)
	s.Require().NoError(err)
}

func TestLedgerIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LedgerIntegrationTestSuite))
}

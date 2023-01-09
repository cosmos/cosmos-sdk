package authz

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/simapp"
	"github.com/cosmos/gogoproto/proto"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	authzclitestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

type fixture struct {
	cfg     network.Config
	network *network.Network
	grantee []sdk.AccAddress
}

func NewFixture(cfg network.Config) *fixture {
	return &fixture{cfg: cfg}
}

func initFixture(t *testing.T) *fixture {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1

	f := NewFixture(cfg)

	t.Log("setting up e2e test suite")

	var err error
	f.network, err = network.New(t, t.TempDir(), f.cfg)
	assert.NilError(t, err)

	val := f.network.Validators[0]
	f.grantee = make([]sdk.AccAddress, 6)

	// Send some funds to the new account.
	// Create new account in the keyring.
	f.grantee[0] = f.createAccount(t, "grantee1")
	f.msgSendExec(t, f.grantee[0])

	// create a proposal with deposit
	_, err = govtestutil.MsgSubmitLegacyProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 1", "Where is the title!?", govv1beta1.ProposalTypeText,
		fmt.Sprintf("--%s=%s", govcli.FlagDeposit, sdk.NewCoin(f.cfg.BondDenom, govv1.DefaultMinDepositTokens).String()))
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	// Create new account in the keyring.
	f.grantee[1] = f.createAccount(t, "grantee2")
	// Send some funds to the new account.
	f.msgSendExec(t, f.grantee[1])

	// grant send authorization to grantee2
	out, err := authzclitestutil.CreateGrant(val.ClientCtx, []string{
		f.grantee[1].String(),
		"send",
		fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
	})
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())
	var response sdk.TxResponse
	assert.NilError(t, val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, response.TxHash, 0))

	// Create new account in the keyring.
	f.grantee[2] = f.createAccount(t, "grantee3")

	// grant send authorization to grantee3
	_, err = authzclitestutil.CreateGrant(val.ClientCtx, []string{
		f.grantee[2].String(),
		"send",
		fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
	})
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	// Create new accounts in the keyring.
	f.grantee[3] = f.createAccount(t, "grantee4")
	f.msgSendExec(t, f.grantee[3])

	f.grantee[4] = f.createAccount(t, "grantee5")
	f.grantee[5] = f.createAccount(t, "grantee6")

	// grant send authorization with allow list to grantee4
	out, err = authzclitestutil.CreateGrant(val.ClientCtx,
		[]string{
			f.grantee[3].String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			fmt.Sprintf("--%s=%s", cli.FlagAllowList, f.grantee[4]),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	assert.NilError(t, val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, response.TxHash, 0))

	return f
}

func (f *fixture) createAccount(t *testing.T, uid string) sdk.AccAddress {
	val := f.network.Validators[0]
	// Create new account in the keyring.
	k, _, err := val.ClientCtx.Keyring.NewMnemonic(uid, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	assert.NilError(t, err)

	addr, err := k.GetAddress()
	assert.NilError(t, err)

	return addr
}

func (f *fixture) msgSendExec(t *testing.T, grantee sdk.AccAddress) {
	val := f.network.Validators[0]
	// Send some funds to the new account.
	out, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		grantee,
		sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(200))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	assert.NilError(t, err)
	assert.DeepEqual(t, strings.Contains(out.String(), `"code":0`), true)
	assert.NilError(t, f.network.WaitForNextBlock())
}

func (f *fixture) TearDownSuite(t *testing.T) {
	t.Log("tearing down e2e test suite")
	f.network.Cleanup()
}

var (
	typeMsgSend           = bank.SendAuthorization{}.MsgTypeURL()
	typeMsgVote           = sdk.MsgTypeURL(&govv1.MsgVote{})
	typeMsgSubmitProposal = sdk.MsgTypeURL(&govv1.MsgSubmitProposal{})
)

func TestCLITxGrantAuthorization(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)

	val := f.network.Validators[0]
	grantee := f.grantee[0]

	twoHours := time.Now().Add(time.Minute * 120).Unix()
	pastHour := time.Now().Add(-time.Minute * 60).Unix()

	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		expErrMsg    string
	}{
		{
			"Invalid granter Address",
			[]string{
				"grantee_addr",
				"send",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "granter"),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			},
			0,
			true,
			`granter.info: key not found`,
		},
		{
			"Invalid grantee Address",
			[]string{
				"grantee_addr",
				"send",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			},
			0,
			true,
			`decoding bech32 failed: invalid separator index -1`,
		},
		{
			"Invalid expiration time",
			[]string{
				grantee.String(),
				"send",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagBroadcastMode),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, pastHour),
			},
			0,
			true,
			"EOF",
		},
		{
			"fail with error invalid msg-type",
			[]string{
				grantee.String(),
				"generic",
				fmt.Sprintf("--%s=invalid-msg-type", cli.FlagMsgType),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			},
			0x1d,
			false,
			"",
		},
		{
			"failed with error both validators not allowed",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", cli.FlagDenyValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			true,
			`cannot set both allowed & deny list`,
		},
		{
			"invalid bond denom for tx delegate authorization allowed validators",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100xyz", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			true,
			`invalid denom xyz; coin denom should match the current bond denom stake`,
		},
		{
			"invalid bond denom for tx delegate authorization deny validators",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100xyz", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagDenyValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			true,
			`invalid denom xyz; coin denom should match the current bond denom stake`,
		},
		{
			"invalid bond denom for tx undelegate authorization",
			[]string{
				grantee.String(),
				"unbond",
				fmt.Sprintf("--%s=100xyz", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			true,
			`invalid denom xyz; coin denom should match the current bond denom stake`,
		},
		{
			"invalid bond denon for tx redelegate authorization",
			[]string{
				grantee.String(),
				"redelegate",
				fmt.Sprintf("--%s=100xyz", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			true,
			`invalid denom xyz; coin denom should match the current bond denom stake`,
		},
		{
			"invalid decimal coin expression with more than single coin",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100stake,20xyz", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			true,
			"invalid decimal coin expression",
		},
		{
			"valid tx delegate authorization allowed validators",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			false,
			"",
		},
		{
			"valid tx delegate authorization deny validators",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagDenyValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			false,
			"",
		},
		{
			"valid tx undelegate authorization",
			[]string{
				grantee.String(),
				"unbond",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			false,
			"",
		},
		{
			"valid tx redelegate authorization",
			[]string{
				grantee.String(),
				"redelegate",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			false,
			"",
		},
		{
			"Valid tx send authorization",
			[]string{
				grantee.String(),
				"send",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			false,
			"",
		},
		{
			"Valid tx send authorization with allow list",
			[]string{
				grantee.String(),
				"send",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", cli.FlagAllowList, f.grantee[1]),
			},
			0,
			false,
			"",
		},
		{
			"Invalid tx send authorization with duplicate allow list",
			[]string{
				grantee.String(),
				"send",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", cli.FlagAllowList, fmt.Sprintf("%s,%s", f.grantee[1], f.grantee[1])),
			},
			0,
			true,
			"duplicate entry",
		},
		{
			"Valid tx generic authorization",
			[]string{
				grantee.String(),
				"generic",
				fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			false,
			"",
		},
		{
			"fail when granter = grantee",
			[]string{
				grantee.String(),
				"generic",
				fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			0,
			true,
			"grantee and granter should be different",
		},
		{
			"Valid tx with amino",
			[]string{
				grantee.String(),
				"generic",
				fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
			0,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := authzclitestutil.CreateGrant(val.ClientCtx, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				var txResp sdk.TxResponse
				assert.NilError(t, err)
				assert.NilError(t, val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
				assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func execDelegate(val *network.Validator, args []string) (testutil.BufferWriter, error) {
	cmd := stakingcli.NewDelegateCmd()
	clientCtx := val.ClientCtx
	return clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
}

func TestCmdRevokeAuthorizations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]

	grantee := f.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	// send-authorization
	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	// generic-authorization
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	// generic-authorization used for amino testing
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgSubmitProposal),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	testCases := []struct {
		name         string
		args         []string
		respType     proto.Message
		expectedCode uint32
		expectErr    bool
	}{
		{
			"invalid grantee address",
			[]string{
				"invalid grantee",
				typeMsgSend,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			nil,
			0,
			true,
		},
		{
			"invalid granter address",
			[]string{
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "granter"),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			nil,
			0,
			true,
		},
		{
			"Valid tx send authorization",
			[]string{
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
		{
			"Valid tx generic authorization",
			[]string{
				grantee.String(),
				typeMsgVote,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
		{
			"Valid tx with amino",
			[]string{
				grantee.String(),
				typeMsgSubmitProposal,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.NewCmdRevokeAuthorization()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
			} else {
				assert.NilError(t, err)
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func TestExecAuthorizationWithExpiration(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]
	grantee := f.grantee[0]
	tenSeconds := time.Now().Add(time.Second * time.Duration(10)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, tenSeconds),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	// msg vote
	voteTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.gov.v1.MsgVote","proposal_id":"1","voter":"%s","option":"VOTE_OPTION_YES"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String())
	execMsg := testutil.WriteToNewTempFile(t, voteTx)
	defer execMsg.Close()

	// waiting for authorization to expires
	time.Sleep(12 * time.Second)

	cmd := cli.NewCmdExecAuthorization()
	clientCtx := val.ClientCtx

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
		execMsg.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	})
	assert.NilError(t, err)
	var response sdk.TxResponse
	assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	assert.NilError(t, clitestutil.CheckTxCode(f.network, clientCtx, response.TxHash, authz.ErrNoAuthorizationFound.ABCICode()))
}

func TestNewExecGenericAuthorized(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]
	grantee := f.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	// msg vote
	voteTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.gov.v1.MsgVote","proposal_id":"1","voter":"%s","option":"VOTE_OPTION_YES"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String())
	execMsg := testutil.WriteToNewTempFile(t, voteTx)
	defer execMsg.Close()

	testCases := []struct {
		name         string
		args         []string
		respType     proto.Message
		expectedCode uint32
		expectErr    bool
	}{
		{
			"fail invalid grantee",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "grantee"),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			nil,
			0,
			true,
		},
		{
			"fail invalid json path",
			[]string{
				"/invalid/file.txt",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			},
			nil,
			0,
			true,
		},
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			&sdk.TxResponse{},
			0,
			false,
		},
		{
			"valid tx with amino",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.NewCmdExecAuthorization()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
			} else {
				assert.NilError(t, err)
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func TestNewExecGrantAuthorized(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)

	val := f.network.Validators[0]
	grantee := f.grantee[0]
	grantee1 := f.grantee[2]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=12%stoken", cli.FlagSpendLimit, val.Moniker),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	tokens := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(12)),
	)
	normalGeneratedTx, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		grantee,
		tokens,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	assert.NilError(t, err)
	execMsg := testutil.WriteToNewTempFile(t, normalGeneratedTx.String())
	defer execMsg.Close()
	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		expectErrMsg string
	}{
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"error over grantee doesn't exist on chain",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee1.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			true,
			"insufficient funds", // earlier the error was account not found here.
		},
		{
			"error over spent",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			authz.ErrNoAuthorizationFound.ABCICode(),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.NewCmdExecAuthorization()
			clientCtx := val.ClientCtx

			var response sdk.TxResponse
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			switch {
			case tc.expectErrMsg != "":
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				assert.Equal(t, strings.Contains(response.RawLog, tc.expectErrMsg), true)

			case tc.expectErr:
				assert.ErrorContains(t, err, "")

			default:
				assert.NilError(t, err)
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}
}

func TestExecSendAuthzWithAllowList(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]
	grantee := f.grantee[3]
	allowedAddr := f.grantee[4]
	notAllowedAddr := f.grantee[5]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
			fmt.Sprintf("--%s=%s", cli.FlagAllowList, allowedAddr),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	tokens := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(12)),
	)

	validGeneratedTx, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		allowedAddr,
		tokens,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	assert.NilError(t, err)
	execMsg := testutil.WriteToNewTempFile(t, validGeneratedTx.String())
	defer execMsg.Close()

	invalidGeneratedTx, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		notAllowedAddr,
		tokens,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	assert.NilError(t, err)
	execMsg1 := testutil.WriteToNewTempFile(t, invalidGeneratedTx.String())
	defer execMsg1.Close()

	// test sending to allowed address
	args := []string{
		execMsg.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
	var response sdk.TxResponse
	cmd := cli.NewCmdExecAuthorization()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	assert.NilError(t, err)
	assert.NilError(t, val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	assert.NilError(t, f.network.WaitForNextBlock())

	// test sending to not allowed address
	args = []string{
		execMsg1.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	assert.NilError(t, err)
	assert.NilError(t, val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	assert.NilError(t, f.network.WaitForNextBlock())

	// query tx and check result
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, authcli.QueryTxCmd(), []string{response.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
	assert.NilError(t, err)
	assert.Equal(t, strings.Contains(out.String(), fmt.Sprintf("cannot send to %s address", notAllowedAddr)), true)
}

func TestExecDelegateAuthorization(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]
	grantee := f.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	tokens := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	delegateTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgDelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg := testutil.WriteToNewTempFile(t, delegateTx)
	defer execMsg.Close()

	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		errMsg       string
	}{
		{
			"valid txn: (delegate half tokens)",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"valid txn: (delegate remaining half tokens)",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"failed with error no authorization found",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			authz.ErrNoAuthorizationFound.ABCICode(),
			false,
			authz.ErrNoAuthorizationFound.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.NewCmdExecAuthorization()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
				assert.Equal(t, strings.Contains(err.Error(), tc.errMsg), true)
			} else {
				var response sdk.TxResponse
				assert.NilError(t, err)
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}

	// test delegate no spend-limit
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	tokens = sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	delegateTx = fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgDelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg = testutil.WriteToNewTempFile(t, delegateTx)
	defer execMsg.Close()

	testCases = []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		errMsg       string
	}{
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.NewCmdExecAuthorization()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
				assert.Equal(t, strings.Contains(err.Error(), tc.errMsg), true)
			} else {
				var response sdk.TxResponse
				assert.NilError(t, err)
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}

	// test delegating to denied validator
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagDenyValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	args := []string{
		execMsg.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
	cmd := cli.NewCmdExecAuthorization()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	var response sdk.TxResponse
	assert.NilError(t, val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())

	// query tx and check result
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, authcli.QueryTxCmd(), []string{response.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
	assert.NilError(t, err)
	assert.Equal(t, strings.Contains(out.String(), fmt.Sprintf("cannot delegate/undelegate to %s validator", val.ValAddress.String())), true)
}

func TestExecUndelegateAuthorization(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]
	grantee := f.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	// granting undelegate msg authorization
	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"unbond",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	// delegating stakes to validator
	_, err = execDelegate(
		val,
		[]string{
			val.ValAddress.String(),
			"100stake",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)

	tokens := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	undelegateTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgUndelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg := testutil.WriteToNewTempFile(t, undelegateTx)
	defer execMsg.Close()

	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		errMsg       string
	}{
		{
			"valid txn: (undelegate half tokens)",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"valid txn: (undelegate remaining half tokens)",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"failed with error no authorization found",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			authz.ErrNoAuthorizationFound.ABCICode(),
			false,
			authz.ErrNoAuthorizationFound.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.NewCmdExecAuthorization()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
				assert.Equal(t, strings.Contains(err.Error(), tc.errMsg), true)
			} else {
				var response sdk.TxResponse
				assert.NilError(t, err)
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}

	// grant undelegate authorization without limit
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"unbond",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	tokens = sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	undelegateTx = fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgUndelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg = testutil.WriteToNewTempFile(t, undelegateTx)
	defer execMsg.Close()

	testCases = []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		errMsg       string
	}{
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.NewCmdExecAuthorization()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
				assert.Equal(t, strings.Contains(err.Error(), tc.errMsg), true)
			} else {
				var response sdk.TxResponse
				assert.NilError(t, err)
				assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				assert.NilError(t, clitestutil.CheckTxCode(f.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}
}

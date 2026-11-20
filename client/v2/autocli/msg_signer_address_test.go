package autocli

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// TestMsgSignerAddressNetwork rekeys an account to a new key, then sends
// coins from it with the AutoCLI bank send command signed by the new key. The
// tx passes the ante handler only with --signer-address, which makes the
// account the msg signer and the account whose number and sequence are signed.
func TestMsgSignerAddressNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	cfg, err := network.DefaultConfigWithAppConfig(authtestutil.AppConfig)
	require.NoError(t, err)
	cfg.NumValidators = 1
	var authGen authtypes.GenesisState
	require.NoError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[authtypes.ModuleName], &authGen))
	authGen.Params.PubKeyChangeEnabled = true
	cfg.GenesisState[authtypes.ModuleName], err = cfg.Codec.MarshalJSON(&authGen)
	require.NoError(t, err)

	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	_, err = net.WaitForHeight(1)
	require.NoError(t, err)

	val := net.Validators[0]
	clientCtx := val.ClientCtx
	kr := clientCtx.Keyring
	ac := clientCtx.TxConfig.SigningContext().AddressCodec()
	txFlags := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, math.NewInt(100)))),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 2_000_000),
	}
	requireTxOK := func(out testutil.BufferWriter) {
		t.Helper()
		var res sdk.TxResponse
		require.NoError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())
		require.Equal(t, uint32(0), res.Code, res.RawLog)
		require.NoError(t, clitestutil.CheckTxCode(net, clientCtx, res.TxHash, 0))
	}

	oldRec, _, err := kr.NewMnemonic("autocli-old", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	account, err := oldRec.GetAddress()
	require.NoError(t, err)
	newRec, _, err := kr.NewMnemonic("autocli-new", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	newPk, err := newRec.GetPubKey()
	require.NoError(t, err)
	require.NotEqual(t, account, sdk.AccAddress(newPk.Address()))

	out, err := clitestutil.MsgSendExec(clientCtx, val.Address, account,
		sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 1_000_000)), ac, txFlags...)
	require.NoError(t, err)
	requireTxOK(out)

	// Rekey the account to the new key.
	pkJSON, err := clientCtx.Codec.MarshalInterfaceJSON(newPk)
	require.NoError(t, err)
	out, err = clitestutil.ExecTestCLICmd(clientCtx, authcli.NewSignRekeyProofCmd(), []string{
		account.String(), string(pkJSON),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, "autocli-new"),
	})
	require.NoError(t, err)
	proofFile := testutil.WriteToNewTempFile(t, out.String()).Name()
	out, err = clitestutil.ExecTestCLICmd(clientCtx, authcli.NewChangePubKeyCmd(), append([]string{
		account.String(), string(pkJSON),
		"--proof=" + proofFile,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, "autocli-old"),
	}, txFlags...))
	require.NoError(t, err)
	requireTxOK(out)

	mergedFiles, err := proto.MergedRegistry()
	require.NoError(t, err)
	b := &Builder{
		Builder: flag.Builder{
			TypeResolver:          protoregistry.GlobalTypes,
			FileResolver:          mergedFiles,
			AddressCodec:          ac,
			ValidatorAddressCodec: clientCtx.TxConfig.SigningContext().ValidatorAddressCodec(),
			ConsensusAddressCodec: addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
		},
		GetClientConn: func(*cobra.Command) (grpc.ClientConnInterface, error) {
			return clientCtx, nil
		},
		AddQueryConnFlags: flags.AddQueryFlagsToCmd,
		AddTxConnFlags:    flags.AddTxFlagsToCmd,
	}
	require.NoError(t, b.ValidateAndComplete())
	sendCmd := func() *cobra.Command {
		cmd := topLevelCmd(context.Background(), "bank", "bank tx commands")
		require.NoError(t, b.AddMsgServiceCommands(cmd, bankAutoCLI))
		return cmd
	}

	recipient := sdk.AccAddress("autocli_recipient___")
	sendAmt := sdk.NewInt64Coin(cfg.BondDenom, 1234)
	sendArgs := append([]string{"send", "autocli-new", recipient.String(), sendAmt.String()}, txFlags...)

	// Without --signer-address the signer is the new key's own address,
	// which has no account.
	_, err = clitestutil.ExecTestCLICmd(clientCtx, sendCmd(), sendArgs)
	require.ErrorContains(t, err, "not found")

	out, err = clitestutil.ExecTestCLICmd(clientCtx, sendCmd(), append(sendArgs,
		fmt.Sprintf("--%s=%s", flags.FlagSignerAddress, account.String())))
	require.NoError(t, err)
	requireTxOK(out)

	res, err := banktypes.NewQueryClient(clientCtx).Balance(context.Background(),
		&banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: cfg.BondDenom})
	require.NoError(t, err)
	require.Equal(t, sendAmt.String(), res.Balance.String())
}

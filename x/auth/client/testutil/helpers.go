package testutil

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cli2 "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

func TxSignExec(clientCtx client.Context, from sdk.AccAddress, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--from=%s", from.String()),
		fmt.Sprintf("--%s=%s", flags.FlagHome, strings.Replace(clientCtx.HomeDir, "simd", "simcli", 1)),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
	}

	args = append(args, extraArgs...)

	return callCmd(clientCtx, cli2.GetSignCommand, args)
}

func callCmd(clientCtx client.Context, theCmd func(clientCtx client.Context) *cobra.Command, extraArgs []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	clientCtx = clientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd := theCmd(clientCtx)
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	cmd.SetArgs(extraArgs)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

// TxSign is simcli sign
func TxSign(f *cli.Fixtures, signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign %v --keyring-backend=test --from=%s %v", f.SimdBinary, f.Flags(), signer, fileName)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxBroadcast is simcli tx broadcast
func TxBroadcast(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx broadcast %v %v", f.SimdBinary, f.Flags(), fileName)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxEncode is simcli tx encode
func TxEncode(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx encode %v %v", f.SimdBinary, f.Flags(), fileName)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

func TxValidateSignaturesExec(clientCtx client.Context, filename string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
	}

	return callCmd(clientCtx, cli2.GetValidateSignaturesCommand, args)
}

// TxValidateSignatures is simcli tx validate-signatures
func TxValidateSignatures(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx validate-signatures %v --keyring-backend=test %v", f.SimdBinary,
		f.Flags(), fileName)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxMultisign is simcli tx multisign
func TxMultisign(f *cli.Fixtures, fileName, name string, signaturesFiles []string,
	flags ...string) (bool, string, string) {

	cmd := fmt.Sprintf("%s tx multisign --keyring-backend=test %v %s %s %s", f.SimdBinary, f.Flags(),
		fileName, name, strings.Join(signaturesFiles, " "),
	)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags))
}

func TxSignBatchExec(clientCtx client.Context, from sdk.AccAddress, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--from=%s", from.String()),
		filename,
	}

	args = append(args, extraArgs...)

	return callCmd(clientCtx, cli2.GetSignBatchCommand, args)
}

// TxDecode is simcli tx decode
func TxDecode(f *cli.Fixtures, encodedTx string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx decode %v %v", f.SimdBinary, f.Flags(), encodedTx)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// DONTCOVER

package testutil

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cli2 "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

func TxSignExec(clientCtx client.Context, from fmt.Stringer, filename string, extraArgs ...string) ([]byte, error) {
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

func TxBroadcastExec(clientCtx client.Context, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		filename,
	}

	args = append(args, extraArgs...)

	return callCmd(clientCtx, cli2.GetBroadcastCommand, args)
}

func callCmd(clientCtx client.Context, theCmd func() *cobra.Command, extraArgs []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	clientCtx = clientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd := theCmd()
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	cmd.SetArgs(extraArgs)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func TxEncodeExec(clientCtx client.Context, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		filename,
	}

	args = append(args, extraArgs...)

	return callCmd(clientCtx, cli2.GetEncodeCommand, args)
}

func TxValidateSignaturesExec(clientCtx client.Context, filename string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
	}

	return callCmd(clientCtx, cli2.GetValidateSignaturesCommand, args)
}

func TxMultiSignExec(clientCtx client.Context, from string, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
		from,
	}

	args = append(args, extraArgs...)

	return callCmd(clientCtx, cli2.GetMultiSignCommand, args)
}

func TxSignBatchExec(clientCtx client.Context, from fmt.Stringer, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--from=%s", from.String()),
		filename,
	}

	args = append(args, extraArgs...)

	return callCmd(clientCtx, cli2.GetSignBatchCommand, args)
}

func TxDecodeExec(clientCtx client.Context, encodedTx string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		encodedTx,
	}

	args = append(args, extraArgs...)

	return callCmd(clientCtx, cli2.GetDecodeCommand, args)
}

// DONTCOVER

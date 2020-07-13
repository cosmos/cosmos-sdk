package testutil

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/x/auth/client/cli"
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

	return client.CallCliCmd(clientCtx, cli.GetSignCommand, args)
}

func TxBroadcastExec(clientCtx client.Context, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		filename,
	}

	args = append(args, extraArgs...)

	return client.CallCliCmd(clientCtx, cli.GetBroadcastCommand, args)
}

func TxEncodeExec(clientCtx client.Context, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		filename,
	}

	args = append(args, extraArgs...)

	return client.CallCliCmd(clientCtx, cli.GetEncodeCommand, args)
}

func TxValidateSignaturesExec(clientCtx client.Context, filename string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
	}

	return client.CallCliCmd(clientCtx, cli.GetValidateSignaturesCommand, args)
}

func TxMultiSignExec(clientCtx client.Context, from string, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, clientCtx.ChainID),
		filename,
		from,
	}

	args = append(args, extraArgs...)

	return client.CallCliCmd(clientCtx, cli.GetMultiSignCommand, args)
}

func TxSignBatchExec(clientCtx client.Context, from fmt.Stringer, filename string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		fmt.Sprintf("--from=%s", from.String()),
		filename,
	}

	args = append(args, extraArgs...)

	return client.CallCliCmd(clientCtx, cli.GetSignBatchCommand, args)
}

func TxDecodeExec(clientCtx client.Context, encodedTx string, extraArgs ...string) ([]byte, error) {
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
		encodedTx,
	}

	args = append(args, extraArgs...)

	return client.CallCliCmd(clientCtx, cli.GetDecodeCommand, args)
}

// DONTCOVER

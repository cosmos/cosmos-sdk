package testutil

import (
	"fmt"
	"strings"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests/cli"
)

// TxSign is simcli sign
func TxSign(f *cli.Fixtures, signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign %v --keyring-backend=test --from=%s %v", f.SimcliBinary, f.Flags(), signer, fileName)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxBroadcast is simcli tx broadcast
func TxBroadcast(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx broadcast %v %v", f.SimcliBinary, f.Flags(), fileName)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxEncode is simcli tx encode
func TxEncode(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx encode %v %v", f.SimcliBinary, f.Flags(), fileName)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxValidateSignatures is simcli tx validate-signatures
func TxValidateSignatures(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx validate-signatures %v --keyring-backend=test %v", f.SimcliBinary,
		f.Flags(), fileName)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxMultisign is simcli tx multisign
func TxMultisign(f *cli.Fixtures, fileName, name string, signaturesFiles []string,
	flags ...string) (bool, string, string) {

	cmd := fmt.Sprintf("%s tx multisign --keyring-backend=test %v %s %s %s", f.SimcliBinary, f.Flags(),
		fileName, name, strings.Join(signaturesFiles, " "),
	)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags))
}

func TxSignBatch(f *cli.Fixtures, signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign-batch %v --keyring-backend=test --from=%s %v", f.SimcliBinary, f.Flags(), signer, fileName)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// DONTCOVER

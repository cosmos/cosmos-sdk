package cli

import (
	"fmt"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	cli "github.com/cosmos/cosmos-sdk/tests/cli"
	"strings"
)

// TxSign is gaiacli tx sign
func TxSign(f *cli.Fixtures, signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign %v --keyring-backend=test --from=%s %v", f.SimcliBinary,
		f.Flags(), signer, fileName)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxBroadcast is gaiacli tx broadcast
func TxBroadcast(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx broadcast %v %v", f.SimcliBinary, f.Flags(), fileName)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxEncode is gaiacli tx encode
func TxEncode(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx encode %v %v", f.SimcliBinary, f.Flags(), fileName)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxMultisign is gaiacli tx multisign
func TxMultisign(f *cli.Fixtures, fileName, name string, signaturesFiles []string,
	flags ...string) (bool, string, string) {

	cmd := fmt.Sprintf("%s tx multisign --keyring-backend=test %v %s %s %s", f.SimcliBinary, f.Flags(),
		fileName, name, strings.Join(signaturesFiles, " "),
	)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags))
}

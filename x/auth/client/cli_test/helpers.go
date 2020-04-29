package cli

import (
	"fmt"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	"strings"
)

// TxSign is gaiacli tx sign
func TxSign(f *helpers.Fixtures, signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign %v --keyring-backend=test --from=%s %v", f.SimcliBinary,
		f.Flags(), signer, fileName)
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxBroadcast is gaiacli tx broadcast
func TxBroadcast(f *helpers.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx broadcast %v %v", f.SimcliBinary, f.Flags(), fileName)
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxEncode is gaiacli tx encode
func TxEncode(f *helpers.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx encode %v %v", f.SimcliBinary, f.Flags(), fileName)
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxMultisign is gaiacli tx multisign
func TxMultisign(f *helpers.Fixtures, fileName, name string, signaturesFiles []string,
	flags ...string) (bool, string, string) {

	cmd := fmt.Sprintf("%s tx multisign --keyring-backend=test %v %s %s %s", f.SimcliBinary, f.Flags(),
		fileName, name, strings.Join(signaturesFiles, " "),
	)
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags))
}

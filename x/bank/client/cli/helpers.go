package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/cli_test/helpers"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

// TxSend is gaiacli tx send
func TxSend(f *helpers.Fixtures, from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx send --keyring-backend=test %s %s %s %v", f.SimcliBinary, from,
		to, amount, f.Flags())
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
